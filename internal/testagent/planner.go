package testagent

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/rocky-ads/site/internal/browserclient"
	"github.com/rocky-ads/site/internal/service/grok"
)

// PlannedAction is a Grok-chosen action.
type PlannedAction struct {
	Action string            `json:"action"`
	Path   string            `json:"path,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
	Reason string            `json:"reason,omitempty"`
	Wait   int               `json:"wait_seconds,omitempty"`
}

const plannerConvID = "test-agent-planner"

const plannerSystemPrompt = `You are simulating a real user on a classified ads website.
Other users on the site are real people — never mention agents, bots, or testing.
Choose ONE next action based only on the affordances provided.
Output ONLY valid JSON with no markdown fences.

Allowed action types:
- {"action":"get","path":"/some/page","reason":"..."} — full page navigation (use page links only)
- {"action":"click","path":"/some/htmx","reason":"..."} — in-page HTMX control (tabs, swaps)
- {"action":"post_form","path":"/api/login","fields":{"username":"x","password":"y"},"reason":"..."}
- {"action":"wait","wait_seconds":30,"reason":"..."}
- {"action":"noop","reason":"..."}

Rules:
- use get only for paths listed under "Page links"
- use click only for paths listed under "HTMX controls"
- for post_form, fields must match form field names from affordances
- prefer natural browsing; do not repeat the same action more than twice in a row
- if you see an error, try a different approach
- do not repeatedly open or close location/search modals — set location once or move on
- if not logged in, use /login or /register when you need to bookmark, message, or post ads
- if already logged in, avoid /register (it logs you out); use /login only if session expired
- never wait on a page with no links; navigate to / or click an ad
- never use /admin/ paths — test agents are not admins
- on my-ads pages, switch tabs with click not get
- do not cycle through my-ads tabs repeatedly; after visiting a tab once, browse ads or create a new ad
- do not noop more than twice on the same page; browse elsewhere instead
- do not change password on settings pages; use notifications and other settings only
- for post_form on /auth/ad/new include title and description; omit images and unknown fields
- check current_category in affordances; use click on /api/category/.../switch to change category (category-select only opens the picker)
- if conversation.messages shows a new reply from the other person, reply with post_form to the conversation send path
- if has_unread_messages is true, go to /auth/user/messages and open an unread conversation before noop or wait
- messages with from_self:true are yours — never reply to your own messages; only reply when the latest message has from_self:false
- send at most one message per turn; after you send, wait for the other person before sending again
- never use click on /send paths; use post_form with a content field to send messages`

// Plan asks Grok for the next action.
func Plan(persona Persona, page browserclient.PageAffordances,
	recent []JournalEntry, loggedIn bool, username, phone string) (PlannedAction, error) {
	linkPaths := sortedKeys(browserclient.PageLinkPaths(page))
	htmxPaths := sortedKeys(browserclient.HTMXPaths(page))

	affJSON, _ := json.Marshal(page)
	recentJSON, _ := json.Marshal(summarizeRecent(recent, 8))
	authState := authPrompt(loggedIn, username, phone)

	userPrompt := fmt.Sprintf(`Persona: %s — %s

%s

Current page affordances:
%s

Page links (use action get):
%s

HTMX controls (use action click):
%s

Recent journal (newest last):
%s

What is your next action?`, persona.Name, persona.Description, authState,
		string(affJSON), strings.Join(linkPaths, "\n"), strings.Join(htmxPaths, "\n"),
		string(recentJSON))

	resp, err := grok.CallGrokConv(plannerSystemPrompt, userPrompt, plannerConvID)
	if err != nil {
		return PlannedAction{}, err
	}

	resp = strings.TrimSpace(resp)
	resp = strings.TrimPrefix(resp, "```json")
	resp = strings.TrimPrefix(resp, "```")
	resp = strings.TrimSuffix(resp, "```")
	resp = strings.TrimSpace(resp)

	var act PlannedAction
	if err := json.Unmarshal([]byte(resp), &act); err != nil {
		return PlannedAction{}, fmt.Errorf("parse plan: %w (raw: %s)", err, resp)
	}
	return act, nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for p := range m {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func authPrompt(loggedIn bool, username, phone string) string {
	if loggedIn {
		return fmt.Sprintf("Auth: logged in as %s.", username)
	}
	return fmt.Sprintf(
		"Auth: not logged in. Your intended username is %s and phone is %s — "+
			"register or log in when needed.",
		username, phone)
}

func summarizeRecent(entries []JournalEntry, n int) []JournalEntry {
	if len(entries) <= n {
		return entries
	}
	return entries[len(entries)-n:]
}

func conversationIDFromSendOrZero(path string) int {
	id, ok := browserclient.ConversationIDFromSendPath(path)
	if !ok {
		return 0
	}
	return id
}

// ValidateAction checks the plan against allowed paths.
func ValidateAction(act PlannedAction, page browserclient.PageAffordances,
	loggedIn bool) error {
	if act.Path != "" && browserclient.IsBlockedAgentPath(act.Path) {
		return fmt.Errorf("path not allowed for test agents: %s", act.Path)
	}
	switch act.Action {
	case "wait", "noop":
		return nil
	case "get":
		if act.Path == "" {
			return fmt.Errorf("missing path")
		}
		if !strings.HasPrefix(act.Path, "/") {
			return fmt.Errorf("path must be relative: %s", act.Path)
		}
		if browserclient.PageLinkPaths(page)[act.Path] {
			return nil
		}
		if loggedIn && browserclient.IsAuthNavPath(act.Path) &&
			!browserclient.HTMXPaths(page)[act.Path] {
			return nil
		}
		return fmt.Errorf("path not a page link (use click for HTMX): %s", act.Path)
	case "click":
		if act.Path == "" {
			return fmt.Errorf("missing path")
		}
		if browserclient.IsMessageSendPath(act.Path) {
			return fmt.Errorf("use post_form to send messages, not click: %s", act.Path)
		}
		if browserclient.HTMXPaths(page)[act.Path] {
			return nil
		}
		return fmt.Errorf("path not an HTMX control: %s", act.Path)
	case "post_form":
		if act.Path == "" {
			return fmt.Errorf("missing path")
		}
		if !strings.HasPrefix(act.Path, "/") {
			return fmt.Errorf("path must be relative: %s", act.Path)
		}
		if browserclient.IsBlockedAgentForm(act.Path) {
			return fmt.Errorf("form not allowed for test agents: %s", act.Path)
		}
		if browserclient.IsMessageSendPath(act.Path) &&
			!browserclient.AllowsMessageSend(act.Path, page.Conversation) {
			return fmt.Errorf("cannot send message now: wait for the other person to reply")
		}
		if browserclient.IsMessageSendPath(act.Path) &&
			browserclient.PageHasOpenConversationForm(page, conversationIDFromSendOrZero(act.Path)) {
			return nil
		}
		allowed := page.AllowedPaths()
		if allowed[act.Path] {
			return nil
		}
		return fmt.Errorf("path not in affordances: %s", act.Path)
	default:
		return fmt.Errorf("unknown action: %s", act.Action)
	}
}

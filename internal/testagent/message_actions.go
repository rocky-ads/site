package testagent

import (
	"fmt"
	"strings"

	"github.com/rocky-ads/site/internal/browserclient"
)

func replyReadyAction(page browserclient.PageAffordances, inbox *inbox,
	persona Persona) (PlannedAction, bool) {
	var conv browserclient.ConversationSnapshot
	if page.Conversation != nil && page.Conversation.AwaitingReply() {
		conv = *page.Conversation
	} else if c := inbox.awaitingReplySnapshot(); c.ID != 0 && c.AwaitingReply() {
		conv = c
	} else {
		return PlannedAction{}, false
	}
	if !browserclient.PageHasOpenConversationForm(page, conv.ID) {
		return PlannedAction{}, false
	}
	return PlannedAction{
		Action: "post_form",
		Path:   fmt.Sprintf("/auth/conversation/%d/send", conv.ID),
		Fields: map[string]string{"content": defaultReplyContent(persona)},
		Reason: "reply to message",
	}, true
}

func conversationClickLoopDetected(entries []JournalEntry, openPath string) bool {
	if openPath == "" {
		openPath = repeatedConversationClickPath(entries)
	}
	if openPath == "" {
		return false
	}
	prefix := "CLICK " + openPath
	count := 0
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Phase != PhaseAct || e.Action != prefix {
			continue
		}
		count++
		if count >= 3 {
			return true
		}
	}
	return false
}

func repeatedConversationClickPath(entries []JournalEntry) string {
	counts := map[string]int{}
	var best string
	bestN := 0
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Phase != PhaseAct {
			continue
		}
		path := pathFromAction(e.Action)
		if !strings.HasPrefix(path, "/auth/conversation/") ||
			strings.Contains(path, "/send") {
			continue
		}
		counts[path]++
		if counts[path] > bestN {
			bestN = counts[path]
			best = path
		}
	}
	if bestN >= 3 {
		return best
	}
	return ""
}

func defaultReplyContent(persona Persona) string {
	switch persona.Name {
	case "car_seller":
		return "Yes, it's still available. Happy to answer any questions."
	case "negotiator":
		return "Thanks for reaching out. What price were you thinking?"
	case "messenger", "cross_traffic":
		return "Thanks for getting back to me. I'm still interested."
	default:
		return "Thanks for the message. I'm still interested."
	}
}

func defaultMessageContent(persona Persona, path string) string {
	if browserclient.IsConversationSendPath(path) {
		return defaultReplyContent(persona)
	}
	switch persona.Name {
	case "messenger", "cross_traffic", "bike_buyer":
		return "Hi, is this still available?"
	case "negotiator":
		return "Interested — is the price negotiable?"
	case "adversarial":
		return "Hi, quick question about this listing."
	default:
		return "Hi, I'm interested in this listing."
	}
}

func ensureMessageFields(act PlannedAction, persona Persona) PlannedAction {
	if !browserclient.IsMessageSendPath(act.Path) {
		return act
	}
	if act.Fields == nil {
		act.Fields = map[string]string{}
	}
	if strings.TrimSpace(act.Fields["content"]) == "" {
		act.Fields["content"] = defaultMessageContent(persona, act.Path)
	}
	return act
}

func (a *Agent) recoverConversationSend(openPath string) (PlannedAction, bool) {
	convID, ok := browserclient.ConversationIDFromOpenPath(openPath)
	if !ok {
		return PlannedAction{}, false
	}
	if err := a.ensureConversationForm(convID); err != nil {
		return PlannedAction{}, false
	}
	obs, _ := a.client.Observe()
	a.syncInbox(obs)
	page := obs.Page
	if act, ok := replyReadyAction(page, a.inbox, a.Persona); ok {
		act.Reason = "reply after conversation open loop"
		return act, true
	}
	return PlannedAction{}, false
}

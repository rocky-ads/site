package browserclient

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ConversationMessage is one line in a conversation modal timeline.
type ConversationMessage struct {
	Text     string `json:"text"`
	FromSelf bool   `json:"from_self"`
}

// ConversationSnapshot is parsed message state from a conversation modal.
type ConversationSnapshot struct {
	ID       int                   `json:"id,omitempty"`
	Messages []ConversationMessage `json:"messages,omitempty"`
}

var conversationMessagesID = regexp.MustCompile(`^conversation-(\d+)-messages$`)

func parseConversationMessages(messagesDiv *goquery.Selection) []ConversationMessage {
	var messages []ConversationMessage
	messagesDiv.Children().Each(func(_ int, row *goquery.Selection) {
		class, _ := row.Attr("class")
		if !strings.Contains(class, "justify-end") && !strings.Contains(class, "justify-start") {
			return
		}
		text := strings.TrimSpace(row.Find("div.max-w-\\[70\\%\\] > div.px-4").First().Text())
		if text == "" || text == "Egg thrown" {
			return
		}
		messages = append(messages, ConversationMessage{
			Text:     text,
			FromSelf: strings.Contains(class, "justify-end"),
		})
	})
	return messages
}

func conversationFromMessagesDiv(messagesDiv *goquery.Selection) ConversationSnapshot {
	idAttr, ok := messagesDiv.Attr("id")
	if !ok {
		return ConversationSnapshot{}
	}
	m := conversationMessagesID.FindStringSubmatch(idAttr)
	if len(m) != 2 {
		return ConversationSnapshot{}
	}
	var id int
	if _, err := fmt.Sscanf(m[1], "%d", &id); err != nil {
		return ConversationSnapshot{}
	}
	return ConversationSnapshot{ID: id, Messages: parseConversationMessages(messagesDiv)}
}

// ParseConversationHTML extracts conversation id and messages from modal HTML.
func ParseConversationHTML(html []byte) ConversationSnapshot {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return ConversationSnapshot{}
	}
	messagesDiv := doc.Find("[id^='conversation-'][id$='-messages']").First()
	if messagesDiv.Length() == 0 {
		return ConversationSnapshot{}
	}
	return conversationFromMessagesDiv(messagesDiv)
}

// ParseAllConversationsHTML returns every open conversation modal on the page.
func ParseAllConversationsHTML(html []byte) []ConversationSnapshot {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil
	}
	var out []ConversationSnapshot
	doc.Find("[id^='conversation-'][id$='-messages']").Each(func(_ int, s *goquery.Selection) {
		conv := conversationFromMessagesDiv(s)
		if conv.ID != 0 {
			out = append(out, conv)
		}
	})
	return out
}

// BestConversation prefers a thread where the other party sent the latest message.
func BestConversation(convs []ConversationSnapshot) *ConversationSnapshot {
	for i := range convs {
		if convs[i].AwaitingReply() {
			return &convs[i]
		}
	}
	if len(convs) == 0 {
		return nil
	}
	return &convs[len(convs)-1]
}

// ConversationIDFromOpenPath parses /auth/conversation/{id}.
func ConversationIDFromOpenPath(path string) (int, bool) {
	if !strings.HasPrefix(path, "/auth/conversation/") {
		return 0, false
	}
	rest := strings.TrimPrefix(path, "/auth/conversation/")
	if rest == "" || strings.Contains(rest, "/") {
		return 0, false
	}
	var id int
	if _, err := fmt.Sscanf(rest, "%d", &id); err != nil {
		return 0, false
	}
	return id, true
}

// ConversationIDFromSendPath parses /auth/conversation/{id}/send.
func ConversationIDFromSendPath(path string) (int, bool) {
	if !strings.HasPrefix(path, "/auth/conversation/") || !strings.HasSuffix(path, "/send") {
		return 0, false
	}
	mid := strings.TrimPrefix(path, "/auth/conversation/")
	mid = strings.TrimSuffix(mid, "/send")
	var id int
	if _, err := fmt.Sscanf(mid, "%d", &id); err != nil {
		return 0, false
	}
	return id, true
}

// AdIDFromSendPath parses /auth/ad/{id}/send.
func AdIDFromSendPath(path string) (int, bool) {
	if !strings.HasPrefix(path, "/auth/ad/") || !strings.HasSuffix(path, "/send") {
		return 0, false
	}
	mid := strings.TrimPrefix(path, "/auth/ad/")
	mid = strings.TrimSuffix(mid, "/send")
	var id int
	if _, err := fmt.Sscanf(mid, "%d", &id); err != nil {
		return 0, false
	}
	return id, true
}

// AdNewConversationPath is the HTMX path that opens a message modal on an ad.
func AdNewConversationPath(adID int) string {
	return fmt.Sprintf("/auth/ad/%d/new-conversation", adID)
}

// PageHasOpenConversationForm reports a reply form visible in the DOM.
func PageHasOpenConversationForm(page PageAffordances, id int) bool {
	for _, formID := range page.OpenConversationForms {
		if formID == id {
			return true
		}
	}
	return false
}

// PageHasAdMessageForm reports an open send form for the first message on an ad.
func PageHasAdMessageForm(page PageAffordances, adID int) bool {
	sendPath := fmt.Sprintf("/auth/ad/%d/send", adID)
	for _, p := range page.OpenAdMessageSendPaths {
		if p == sendPath {
			return true
		}
	}
	return false
}

// PageHasConversation reports a parsed conversation modal on the page.
func PageHasConversation(page PageAffordances, id int) bool {
	if page.Conversation != nil && page.Conversation.ID == id {
		return true
	}
	for _, c := range page.Conversations {
		if c.ID == id {
			return true
		}
	}
	return false
}

// AwaitingReply reports whether the other party sent the latest message.
func (c ConversationSnapshot) AwaitingReply() bool {
	if len(c.Messages) == 0 {
		return false
	}
	return !c.Messages[len(c.Messages)-1].FromSelf
}

// ReceivedCount returns messages from the other party.
func (c ConversationSnapshot) ReceivedCount() int {
	n := 0
	for _, m := range c.Messages {
		if !m.FromSelf {
			n++
		}
	}
	return n
}

// EnrichWithConversation adds conversation context; reply form only when awaiting reply.
func EnrichWithConversation(p PageAffordances, conv ConversationSnapshot) PageAffordances {
	if conv.ID == 0 {
		return p
	}
	out := p
	snap := conv
	out.Conversation = &snap
	if !conv.AwaitingReply() {
		return out
	}
	sendPath := fmt.Sprintf("/auth/conversation/%d/send", conv.ID)
	out.Forms = append(out.Forms, Form{
		Action: sendPath,
		Method: "POST",
		Fields: []FormField{{Name: "content", Type: "text"}},
	})
	return out
}

// IsConversationSendPath reports POST targets for replying in a thread.
func IsConversationSendPath(path string) bool {
	return strings.HasPrefix(path, "/auth/conversation/") && strings.HasSuffix(path, "/send")
}

// IsAdSendPath reports POST targets for the first message on an ad.
func IsAdSendPath(path string) bool {
	return strings.HasPrefix(path, "/auth/ad/") && strings.HasSuffix(path, "/send")
}

// IsMessageSendPath reports any conversation message POST target.
func IsMessageSendPath(path string) bool {
	return IsConversationSendPath(path) || IsAdSendPath(path)
}

// AllowsMessageSend reports whether the agent may POST to a message send path.
func AllowsMessageSend(path string, conv *ConversationSnapshot) bool {
	if IsAdSendPath(path) {
		return conv == nil || conv.ID == 0 || len(conv.Messages) == 0
	}
	if IsConversationSendPath(path) {
		if conv == nil || conv.ID == 0 {
			return false
		}
		if path != fmt.Sprintf("/auth/conversation/%d/send", conv.ID) {
			return false
		}
		return conv.AwaitingReply()
	}
	return true
}

// FilterMessageSends removes message POST affordances when sending is not allowed.
func FilterMessageSends(p PageAffordances) PageAffordances {
	conv := p.Conversation
	out := p
	out.Forms = nil
	for _, f := range p.Forms {
		if IsMessageSendPath(f.Action) && !AllowsMessageSend(f.Action, conv) {
			continue
		}
		out.Forms = append(out.Forms, f)
	}
	out.HTMX = nil
	for _, h := range p.HTMX {
		if IsMessageSendPath(h.Path) {
			continue
		}
		out.HTMX = append(out.HTMX, h)
	}
	return out
}

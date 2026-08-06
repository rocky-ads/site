package ui

import (
	"fmt"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// Named SSE events — listeners exist only when the matching UI is mounted,
// so missing-target OOB errors are avoided on other pages.
const (
	SSEEventUnread           = "unread"
	SSEEventConversationList = "conversation-list"
	SSEEventNotifications    = "notifications"
)

func SSEEventConversation(conversationID int) string {
	return fmt.Sprintf("conversation-%d", conversationID)
}

// Hidden sink: hx-swap none so the payload is OOB-only.
func sseOOBSink(event, id string) g.Node {
	return Div(
		ID(id),
		g.Attr("sse-swap", event),
		hx.Swap("none"),
		g.Attr("aria-hidden", "true"),
		g.Attr("style", "display: none;"),
	)
}

func ConversationSSESink(conversationID int) g.Node {
	return sseOOBSink(
		SSEEventConversation(conversationID),
		fmt.Sprintf("conversation-%d-sse", conversationID),
	)
}

func ConversationListSSESink() g.Node {
	return sseOOBSink(SSEEventConversationList, "conversation-list-sse")
}

func UnreadSSESink() g.Node {
	return sseOOBSink(SSEEventUnread, "unread-sse")
}

func NotificationsSSESink() g.Node {
	return sseOOBSink(SSEEventNotifications, "notifications-sse")
}

package ui

import (
	"fmt"
	"time"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func formatMessageTime(t time.Time) string {
	now := time.Now()
	d := now.Sub(t)

	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}

	return t.Format("Jan 2, 2006")
}

func MessageItem(senderID, currentUserID int, content string, createdAt time.Time, loc *time.Location, attrs ...g.Node) g.Node {
	isSent := senderID == currentUserID

	var bubbleClass string
	var containerClass string
	if isSent {
		bubbleClass = "bg-blue-500 text-white rounded-lg rounded-tr-none"
		containerClass = "flex justify-end"
	} else {
		bubbleClass = "bg-zinc-200 dark:bg-zinc-700 text-zinc-900 dark:text-zinc-200 rounded-lg rounded-tl-none"
		containerClass = "flex justify-start"
	}

	localTime := createdAt.In(loc)
	fullTimestamp := localTime.Format("2006-01-02 03:04:05 PM MST")

	allAttrs := append([]g.Node{
		Class(containerClass + " mb-2"),
		Title(fullTimestamp),
	}, attrs...)

	return Div(
		g.Group(allAttrs),
		Div(
			Class("max-w-[70%]"),
			Div(
				Class("px-4 py-2 "+bubbleClass),
				g.Text(content),
			),
			Div(
				Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1 px-1"),
				g.Text(formatMessageTime(localTime)),
			),
		),
	)
}

func ConversationMessages(conversationID, currentUserID int, otherUserName string, messageNodes []g.Node, csrfToken string) g.Node {
	return Div(
		ID(fmt.Sprintf("conversation-%d-messages", conversationID)),
		Class("flex-1 overflow-y-auto p-4 space-y-2"),
		g.Group(messageNodes),
	)
}

func ConversationMessagesSentinel(conversationID int) g.Node {
	return Div(
		ID(fmt.Sprintf("conversation-%d-sentinel", conversationID)),
		g.Attr("style", "display: none;"),
	)
}

func ConversationMessageSwapOOB(conversationID int, messageNode g.Node) g.Node {
	sentinelID := fmt.Sprintf("conversation-%d-sentinel", conversationID)
	return Div(
		hx.SwapOOB(fmt.Sprintf("beforebegin:#%s", sentinelID)),
		messageNode,
	)
}

func ConversationMessagesUpdate(conversationID int, messageNodes []g.Node) g.Node {
	nodesWithSentinel := append(messageNodes, ConversationMessagesSentinel(conversationID))
	return Div(
		ID(fmt.Sprintf("conversation-%d-messages", conversationID)),
		Class("flex-1 overflow-y-auto p-4 space-y-2"),
		g.Group(nodesWithSentinel),
	)
}

func ConversationContentInput(conversationID int, attrs ...g.Node) g.Node {
	allAttrs := []g.Node{
		ID(fmt.Sprintf("conversation-%d-content-input", conversationID)),
		Type("text"),
		Name("content"),
		Placeholder("Type a message..."),
		Required(),
		Class("flex-1 px-4 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-200"),
		g.Attr("onkeydown", "if(event.key==='Enter' && !event.shiftKey) { event.preventDefault(); this.form.requestSubmit(); }"),
	}
	allAttrs = append(allAttrs, attrs...)
	return Input(g.Group(allAttrs))
}

func ConversationForm(conversationID int, csrfToken string) g.Node {
	modalName := fmt.Sprintf("conversation-%d", conversationID)
	attrs := []g.Node{
		ID(fmt.Sprintf("%s-form", modalName)),
		Class("p-4 border-t border-zinc-200 dark:border-zinc-700 flex-shrink-0"),
		hx.Post(fmt.Sprintf("/auth/conversation/%d/send", conversationID)),
		hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
		hx.Include("this"),
		hx.Swap("none"),
	}
	return Form(
		g.Group(attrs),
		Div(
			Class("flex gap-2"),
			ConversationContentInput(conversationID),
			Button(
				Type("submit"),
				Class("px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors"),
				g.Text("Send"),
			),
		),
	)
}

func ConversationModal(conversationID, adID int, adTitle string, ownerID, enquirerID, currentUserID int, otherUserName string, messageNodes []g.Node, csrfToken string) g.Node {
	modalName := fmt.Sprintf("conversation-%d", conversationID)

	return g.Group([]g.Node{
		modalBackdrop(modalName),
		Div(
			ID(modalName+"-modal"),
			Class("fixed inset-0 flex items-center justify-center z-50 p-8 pointer-events-none"),
			Div(
				Class("bg-white dark:bg-zinc-800 rounded-lg w-full max-w-lg shadow-2xl border-2 border-zinc-300 dark:border-zinc-600 flex flex-col pointer-events-auto"),
				Style("max-height: min(80vh, 600px)"),
				Div(
					Class("flex items-start justify-between p-4 border-b border-zinc-200 dark:border-zinc-700 flex-shrink-0"),
					Div(
						Class("flex-1 min-w-0 pr-4"),
						Div(
							Class("text-sm text-zinc-600 dark:text-zinc-400 mb-1"),
							Span(Class("font-semibold"), g.Text("Subject: ")),
							g.Text(adTitle),
						),
						Div(
							Class("text-sm text-zinc-600 dark:text-zinc-400"),
							Span(Class("font-semibold"), g.Text("To: ")),
							g.Text(otherUserName),
						),
					),
					modalClose(modalName),
				),
				Div(
					ID(fmt.Sprintf("%s-messages", modalName)),
					Class("flex-1 overflow-y-auto p-4 space-y-2"),
					g.If(len(messageNodes) == 0,
						Div(
							Class("text-center text-zinc-500 dark:text-zinc-400 py-8"),
							g.Text("No messages yet. Start the conversation!"),
						),
					),
					g.If(len(messageNodes) > 0,
						g.Group(messageNodes),
					),
					ConversationMessagesSentinel(conversationID),
				),
				ConversationForm(conversationID, csrfToken),
			),
		),
	})
}

func ConversationListItem(conversationID, adID int, adTitle, lastMessageContent string, lastMessageAt *time.Time, updatedAt time.Time, otherUserName string) g.Node {
	lastMessagePreview := lastMessageContent
	if len(lastMessagePreview) > 50 {
		lastMessagePreview = lastMessagePreview[:50] + "..."
	}

	var timeStr string
	if lastMessageAt != nil {
		timeStr = formatMessageTime(*lastMessageAt)
	} else {
		timeStr = formatMessageTime(updatedAt)
	}

	return Div(
		Class("border-b border-zinc-200 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-800 cursor-pointer transition-colors"),
		hx.Get(fmt.Sprintf("/auth/conversation/%d", conversationID)),
		hx.Target("body"),
		hx.Swap("beforeend"),
		Div(
			Class("p-4"),
			Div(
				Class("flex items-start justify-between mb-2"),
				A(
					Href(fmt.Sprintf("/ad/%d", adID)),
					Class("text-lg font-semibold text-zinc-900 dark:text-zinc-200 hover:text-blue-600 dark:hover:text-blue-400"),
					g.Text(adTitle),
					g.Attr("onclick", "event.stopPropagation()"),
				),
				Span(
					Class("text-xs text-zinc-500 dark:text-zinc-400 ml-4 flex-shrink-0"),
					g.Text(timeStr),
				),
			),
			Div(
				Class("flex items-center justify-between"),
				Span(
					Class("text-sm text-zinc-600 dark:text-zinc-400"),
					g.Text(fmt.Sprintf("With %s", otherUserName)),
				),
				g.If(lastMessageContent != "",
					Span(
						Class("text-sm text-zinc-500 dark:text-zinc-500 truncate ml-4 max-w-[50%%]"),
						g.Text(lastMessagePreview),
					),
				),
			),
		),
	)
}

func MessagesPage(conversationItems []g.Node) []g.Node {
	var conversationNodes []g.Node
	if len(conversationItems) == 0 {
		conversationNodes = append(conversationNodes,
			Div(
				Class("text-center text-zinc-600 dark:text-zinc-400 py-16"),
				P(g.Text("No conversations yet. Start a conversation by clicking the message button on an ad.")),
			),
		)
	} else {
		conversationNodes = conversationItems
	}

	return []g.Node{
		pageTitle("Messages"),
		Div(
			Class("mt-6"),
			Div(
				Class("bg-white dark:bg-zinc-800 rounded-lg shadow-lg border border-zinc-200 dark:border-zinc-700"),
				g.Group(conversationNodes),
			),
		),
	}
}

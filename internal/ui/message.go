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

func MessageItem(d MessageItemData, attrs ...g.Node) g.Node {
	isSent := d.SenderID == d.CurrentUserID

	var bubbleClass string
	var containerClass string
	if isSent {
		bubbleClass = "bg-blue-500 text-white rounded-lg rounded-tr-none"
		containerClass = "flex justify-end"
	} else {
		bubbleClass = "bg-zinc-200 dark:bg-zinc-700 text-zinc-900 dark:text-zinc-200 rounded-lg rounded-tl-none"
		containerClass = "flex justify-start"
	}

	fullTimestamp := d.CreatedAt.Format("2006-01-02 03:04:05 PM MST")

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
				g.Text(d.Content),
			),
			Div(
				Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1 px-1"),
				g.Text(formatMessageTime(d.CreatedAt)),
			),
		),
	)
}

// RockEventMessage renders a journal entry for a rock throw or unthrow.
func RockEventMessage(d RockEventData, attrs ...g.Node) g.Node {
	isSent := d.ThrowerID == d.CurrentUserID

	var bubbleClass string
	var containerClass string
	if isSent {
		bubbleClass = "bg-blue-500 text-white rounded-lg rounded-tr-none"
		containerClass = "flex justify-end"
	} else {
		bubbleClass = "bg-zinc-200 dark:bg-zinc-700 text-zinc-900 dark:text-zinc-200 rounded-lg rounded-tl-none"
		containerClass = "flex justify-start"
	}

	label := "Rock thrown"
	if d.Kind == RockEventUnthrown {
		label = "Rock unthrown"
	}

	fullTimestamp := d.EventAt.Format("2006-01-02 03:04:05 PM MST")

	allAttrs := append([]g.Node{
		Class(containerClass + " mb-2"),
		Title(fullTimestamp),
	}, attrs...)

	return Div(
		g.Group(allAttrs),
		Div(
			Class("max-w-[70%]"),
			Div(
				Class("px-4 py-2 "+bubbleClass+" flex items-center gap-2"),
				Img(
					Src("/images/rock.svg"),
					Alt(label),
					Class("w-5 h-5 flex-shrink-0"),
				),
				g.Text(label),
			),
			Div(
				Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1 px-1"),
				g.Text(formatMessageTime(d.EventAt)),
			),
		),
	)
}

// MessageTimeline renders chat messages with rock journal entries inserted
// chronologically. Message and rock timestamps must already be in the viewer's
// timezone.
func MessageTimeline(messages []MessageItemData,
	rocks []RockEventData) []g.Node {
	if len(rocks) == 0 {
		nodes := make([]g.Node, len(messages))
		for i, m := range messages {
			nodes[i] = MessageItem(m)
		}
		return nodes
	}

	var nodes []g.Node
	ri, mi := 0, 0
	for ri < len(rocks) || mi < len(messages) {
		useRock := mi >= len(messages) ||
			(ri < len(rocks) &&
				!rocks[ri].EventAt.After(messages[mi].CreatedAt))
		if useRock {
			nodes = append(nodes, RockEventMessage(rocks[ri]))
			ri++
			continue
		}
		nodes = append(nodes, MessageItem(messages[mi]))
		mi++
	}
	return nodes
}

// conversationMessagesSelector is the inner list (append / OOB target).
// The outer #conversation-N-messages scrollport uses column-reverse so
// scrollTop=0 shows the newest messages (go-chat pattern).
func conversationMessagesSelector(conversationID int) string {
	return fmt.Sprintf("#conversation-%d-messages-list", conversationID)
}

func conversationMessagesAppendSwap() string {
	return "beforeend"
}

func ConversationMessageAppendOOB(conversationID int) g.Node {
	return hx.SwapOOB(fmt.Sprintf(
		"beforeend:%s", conversationMessagesSelector(conversationID)))
}

// ConversationEmptyMessageDeleteSwapOOB removes the empty-state placeholder
// when the first message is appended.
func ConversationEmptyMessageDeleteSwapOOB(conversationID int) g.Node {
	return Div(
		ID(fmt.Sprintf("conversation-%d-empty-message", conversationID)),
		hx.SwapOOB("delete"),
	)
}

func ConversationMessagesArea(conversationID int, canPost bool,
	messageNodes []g.Node, extraAttrs ...g.Node) g.Node {
	listAttrs := append([]g.Node{
		ID(fmt.Sprintf("conversation-%d-messages-list", conversationID)),
		Class("flex flex-col space-y-2"),
	}, extraAttrs...)
	return Div(
		ID(fmt.Sprintf("conversation-%d-messages", conversationID)),
		Class("flex-1 min-h-0 overflow-y-auto flex flex-col-reverse p-4"),
		Div(
			g.Group(listAttrs),
			g.If(len(messageNodes) == 0,
				Div(
					ID(fmt.Sprintf("conversation-%d-empty-message", conversationID)),
					Class("text-center text-zinc-500 dark:text-zinc-400 py-8"),
					g.If(canPost,
						g.Text("No messages yet. Start the conversation!"),
					),
					g.If(!canPost,
						g.Text("No messages yet."),
					),
				),
			),
			g.If(len(messageNodes) > 0,
				g.Group(messageNodes),
			),
		),
	)
}

func messageActionsBannerVisible(d ConversationModalData) bool {
	if d.CanPost && d.RockThrowerID != nil {
		return true
	}
	return d.HasThrownRock || d.CanThrowRock
}

func ConversationMessageActionsBanner(d ConversationModalData,
	extraAttrs ...g.Node) g.Node {
	visible := messageActionsBannerVisible(d)
	classes := "w-full flex-shrink-0 flex items-center justify-end gap-3"
	if visible {
		classes += " border-b border-zinc-200 dark:border-zinc-700 px-4 py-2"
	}
	attrs := append([]g.Node{
		ID(fmt.Sprintf("conversation-%d-message-actions", d.ConversationID)),
		Class(classes),
	}, extraAttrs...)
	if !visible {
		return Div(g.Group(attrs))
	}
	return Div(
		g.Group(attrs),
		g.If(d.CanPost && d.RockThrowerID != nil,
			RockOpinionLink(d.ConversationID),
		),
		RockThrowLink(d.ConversationID, d.HasThrownRock,
			d.CanThrowRock, d.CSRFToken),
	)
}

func ConversationMessagesPanel(d ConversationModalData) g.Node {
	return Div(
		Class("flex-1 flex flex-col min-h-0"),
		ConversationSSESink(d.ConversationID),
		ConversationMessageActionsBanner(d),
		ConversationMessagesArea(d.ConversationID, d.CanPost, d.MessageNodes),
	)
}

func ConversationMessagesSelector(conversationID int) string {
	return conversationMessagesSelector(conversationID)
}

// ConversationMessagesAppendSwapForTest exposes the append swap spec for tests.
func ConversationMessagesAppendSwapForTest() string {
	return conversationMessagesAppendSwap()
}

func ConversationMessageActionsSwapOOB(d ConversationModalData) g.Node {
	return ConversationMessageActionsBanner(d, hx.SwapOOB("outerHTML"))
}

func ConversationListItemSwapOOB(conversationID int,
	conversationItemNode g.Node) g.Node {
	itemID := fmt.Sprintf("conversation-item-%d", conversationID)
	return Div(
		hx.SwapOOB("outerHTML"),
		ID(itemID),
		conversationItemNode,
	)
}

func conversationContentInputAttrs(conversationID int) []g.Node {
	return []g.Node{
		ID(fmt.Sprintf("conversation-%d-content-input", conversationID)),
		Type("text"),
		Name("content"),
		Placeholder("Type a message..."),
		Required(),
		Class("flex-1 px-4 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-200"),
		g.Attr("onkeydown", "if(event.key==='Enter' && !event.shiftKey) { event.preventDefault(); this.form.requestSubmit(); }"),
	}
}

func ConversationContentInput(conversationID int, attrs ...g.Node) g.Node {
	allAttrs := append(conversationContentInputAttrs(conversationID), attrs...)
	return Input(g.Group(allAttrs))
}

// ConversationContentInputClearSwapOOB replaces the message input with an empty field after send.
func ConversationContentInputClearSwapOOB(conversationID int) g.Node {
	attrs := append(conversationContentInputAttrs(conversationID), hx.SwapOOB("outerHTML"))
	return Input(g.Group(attrs))
}

func ConversationForm(conversationID, adID int, csrfToken string,
	canPost bool) g.Node {
	modalName := fmt.Sprintf("conversation-%d", conversationID)
	attrs := []g.Node{
		ID(fmt.Sprintf("%s-form", modalName)),
		Class("p-4 border-t border-zinc-200 dark:border-zinc-700 flex-shrink-0"),
	}
	if canPost {
		var postURL string
		if conversationID == 0 {
			// New conversation - use ad-based endpoint
			postURL = fmt.Sprintf("/auth/ad/%d/send", adID)
		} else {
			// Existing conversation - use conversation-based endpoint
			postURL = fmt.Sprintf("/auth/conversation/%d/send", conversationID)
		}
		attrs = append(attrs,
			hx.Post(postURL),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Include("this"),
			hx.Target(conversationMessagesSelector(conversationID)),
			hx.Swap(conversationMessagesAppendSwap()),
		)
	}
	return Form(
		g.Group(attrs),
		Div(
			Class("flex gap-2"),
			g.If(canPost,
				ConversationContentInput(conversationID, Autofocus()),
			),
			g.If(!canPost,
				Input(
					ID(fmt.Sprintf("conversation-%d-content-input", conversationID)),
					Type("text"),
					Name("content"),
					Placeholder("Read-only conversation"),
					Disabled(),
					Class("flex-1 px-4 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md bg-zinc-100 dark:bg-zinc-700 text-zinc-500 dark:text-zinc-400 cursor-not-allowed"),
				),
			),
			g.If(canPost,
				Button(
					Type("submit"),
					Class("px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors"),
					g.Text("Send"),
				),
			),
		),
	)
}

// ConversationModalSwapOOB returns the modal div with hx-swap-oob="outerHTML".
// Used when the first message creates a conversation.
func ConversationModalSwapOOB(d ConversationModalData) g.Node {
	modalName := fmt.Sprintf("conversation-%d", d.ConversationID)
	modalID := modalName + "-modal"
	if d.TargetModalID != "" {
		modalID = d.TargetModalID
	}

	var rockUserID int
	if d.RockThrowerID != nil {
		rockUserID = *d.RockThrowerID
	}

	// Return just the modal div (not the backdrop) with OOB swap
	return Div(
		ID(modalID),
		hx.SwapOOB("outerHTML"),
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
						g.Text(d.AdTitle),
					),
					Div(
						Class("text-sm text-zinc-600 dark:text-zinc-400"),
						Span(Class("font-semibold"), g.Text("From: ")),
						UserRockIcons(d.InquirerID, d.InquirerRockCount),
						UserNameLink(d.InquirerID, d.InquirerName),
						g.Text(", "),
						Span(Class("font-semibold"), g.Text("To: ")),
						UserRockIcons(d.OwnerID, d.OwnerRockCount),
						UserNameLink(d.OwnerID, d.OwnerName),
						g.Text(" (ad owner)"),
						g.If(!d.CanPost && (d.OwnerID != d.CurrentUserID && d.InquirerID != d.CurrentUserID),
							Span(
								Class("ml-2 px-2 py-0.5 bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 rounded text-xs"),
								g.Text("Read-only"),
							),
						),
					),
					g.If(!d.CanPost && d.RockThrowerID != nil,
						Div(
							Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1"),
							g.Text("Rock thrown by: "),
							g.If(rockUserID == d.CurrentUserID,
								Span(Class("text-blue-600 dark:text-blue-400 font-medium"),
									g.If(rockUserID == d.OwnerID, UserNameLink(d.OwnerID, d.OwnerName)),
									g.If(rockUserID == d.InquirerID, UserNameLink(d.InquirerID, d.InquirerName)),
								),
							),
							g.If(rockUserID != d.CurrentUserID,
								Span(Class("text-zinc-700 dark:text-zinc-300 font-medium"),
									g.If(rockUserID == d.OwnerID, UserNameLink(d.OwnerID, d.OwnerName)),
									g.If(rockUserID == d.InquirerID, UserNameLink(d.InquirerID, d.InquirerName)),
								),
							),
						),
					),
				),
				Div(
					ID(fmt.Sprintf("%s-header-actions", modalName)),
					Class("flex items-center gap-2"),
					Button(
						ID(fmt.Sprintf("%s-close-button", modalName)),
						Type("button"),
						Class("bg-white dark:bg-zinc-700 border-2 border-zinc-800 dark:border-zinc-500 rounded-full w-8 h-8 flex items-center justify-center shadow-lg hover:bg-zinc-100 dark:hover:bg-zinc-600 focus:outline-none cursor-pointer"),
						hx.Get(fmt.Sprintf("/api/modal-remove/%s", modalName)),
						hx.Swap("none"),
						Img(
							Src("/images/close.svg"),
							Alt("Close"),
							Class("w-4 h-4 dark:invert"),
						),
					),
				),
			),
			ConversationMessagesPanel(d),
			ConversationForm(
				d.ConversationID, d.AdID, d.CSRFToken, d.CanPost,
			),
		),
	)
}

func ConversationModalWithRock(d ConversationModalData) g.Node {
	modalName := fmt.Sprintf("conversation-%d", d.ConversationID)

	var rockUserID int
	if d.RockThrowerID != nil {
		rockUserID = *d.RockThrowerID
	}

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
							g.Text(d.AdTitle),
						),
						Div(
							Class("text-sm text-zinc-600 dark:text-zinc-400"),
							Span(Class("font-semibold"), g.Text("From: ")),
							UserRockIcons(d.InquirerID, d.InquirerRockCount),
							UserNameLink(d.InquirerID, d.InquirerName),
							g.Text(", "),
							Span(Class("font-semibold"), g.Text("To: ")),
							UserRockIcons(d.OwnerID, d.OwnerRockCount),
							UserNameLink(d.OwnerID, d.OwnerName),
							g.Text(" (ad owner)"),
							g.If(!d.CanPost && (d.OwnerID != d.CurrentUserID && d.InquirerID != d.CurrentUserID),
								Span(
									Class("ml-2 px-2 py-0.5 bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 rounded text-xs"),
									g.Text("Read-only"),
								),
							),
						),
						g.If(!d.CanPost && d.RockThrowerID != nil,
							Div(
								Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1"),
								g.Text("Rock thrown by: "),
								g.If(rockUserID == d.CurrentUserID,
									Span(Class("text-blue-600 dark:text-blue-400 font-medium"),
										g.If(rockUserID == d.OwnerID, UserNameLink(d.OwnerID, d.OwnerName)),
										g.If(rockUserID == d.InquirerID, UserNameLink(d.InquirerID, d.InquirerName)),
									),
								),
								g.If(rockUserID != d.CurrentUserID,
									Span(Class("text-zinc-700 dark:text-zinc-300 font-medium"),
										g.If(rockUserID == d.OwnerID, UserNameLink(d.OwnerID, d.OwnerName)),
										g.If(rockUserID == d.InquirerID, UserNameLink(d.InquirerID, d.InquirerName)),
									),
								),
							),
						),
					),
					Div(
						ID(fmt.Sprintf("%s-header-actions", modalName)),
						Class("flex items-center gap-2"),
						Button(
							ID(fmt.Sprintf("%s-close-button", modalName)),
							Type("button"),
							Class("bg-white dark:bg-zinc-700 border-2 border-zinc-800 dark:border-zinc-500 rounded-full w-8 h-8 flex items-center justify-center shadow-lg hover:bg-zinc-100 dark:hover:bg-zinc-600 focus:outline-none cursor-pointer"),
							hx.Get(fmt.Sprintf("/api/modal-remove/%s", modalName)),
							hx.Swap("none"),
							Img(
								Src("/images/close.svg"),
								Alt("Close"),
								Class("w-4 h-4 dark:invert"),
							),
						),
					),
				),
				ConversationMessagesPanel(d),
				ConversationForm(
					d.ConversationID, d.AdID, d.CSRFToken, d.CanPost,
				),
			),
		),
	})
}

func ConversationListItem(d ConversationListItemData) g.Node {
	lastMessagePreview := d.LastMessageContent
	if len(lastMessagePreview) > 50 {
		lastMessagePreview = lastMessagePreview[:50] + "..."
	}

	var timeStr string
	if d.LastMessageAt != nil {
		timeStr = formatMessageTime(*d.LastMessageAt)
	} else {
		timeStr = formatMessageTime(d.UpdatedAt)
	}

	return Div(
		ID(fmt.Sprintf("conversation-item-%d", d.ConversationID)),
		Class("border-b border-zinc-200 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-800 cursor-pointer transition-colors"),
		hx.Get(fmt.Sprintf("/auth/conversation/%d", d.ConversationID)),
		hx.Target("#page-content"),
		hx.Swap("beforeend"),
		Div(
			Class("p-4"),
			Div(
				Class("flex items-start justify-between mb-2"),
				Div(
					Class("flex items-center gap-2 flex-1 min-w-0"),
					g.If(d.HasUnread,
						Div(
							Class("bg-green-500 rounded-full w-2 h-2 flex-shrink-0"),
						),
					),
					g.If(d.RockCount > 0, StaticRockIcons(d.RockCount)),
					Span(
						Class("text-lg font-semibold text-zinc-900 dark:text-zinc-200"),
						g.Text(d.AdTitle),
					),
				),
				Span(
					Class("text-xs text-zinc-500 dark:text-zinc-400 ml-4 flex-shrink-0"),
					g.Text(timeStr),
				),
			),
			Div(
				Class("flex items-center justify-between"),
				Span(
					Class("text-sm text-zinc-600 dark:text-zinc-400 flex items-center gap-1"),
					g.If(d.InquirerID == d.CurrentUserID,
						g.Group([]g.Node{
							g.Text("To: "),
							UserNameLink(d.OwnerID, d.OtherUserName,
								StaticRockIcons(d.OtherUserRockCount)),
						}),
					),
					g.If(d.OwnerID == d.CurrentUserID,
						g.Group([]g.Node{
							g.Text("From: "),
							UserNameLink(d.InquirerID, d.OtherUserName,
								StaticRockIcons(d.OtherUserRockCount)),
						}),
					),
				),
				g.If(d.LastMessageContent != "",
					Span(
						Class("text-sm text-zinc-500 dark:text-zinc-500 truncate ml-4 max-w-[50%%]"),
						g.Text(lastMessagePreview),
					),
				),
			),
		),
	)
}

func MessagesPage(items []ConversationListItemData) []g.Node {
	var conversationNodes []g.Node
	if len(items) == 0 {
		conversationNodes = append(conversationNodes,
			Div(
				Class("text-center text-zinc-600 dark:text-zinc-400 py-16"),
				P(
					Class("flex items-center justify-center gap-2"),
					g.Text("No conversations yet. Start a conversation by clicking the "),
					Img(
						Src("/images/message.svg"),
						Alt("Message"),
						Class("w-5 h-5 inline-block dark:invert dark:opacity-80"),
					),
					g.Text(" message button on an ad."),
				),
			),
		)
	} else {
		for _, item := range items {
			conversationNodes = append(conversationNodes, ConversationListItem(item))
		}
	}

	return []g.Node{
		pageTitle("Messages"),
		ConversationListSSESink(),
		Div(
			Class("mt-6"),
			Div(
				Class("bg-white dark:bg-zinc-800 rounded-lg shadow-lg border border-zinc-200 dark:border-zinc-700"),
				g.Group(conversationNodes),
			),
		),
	}
}

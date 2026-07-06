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

// RockThrownMessage renders a message-like item showing when a rock was thrown
func RockThrownMessage(d RockEventData) g.Node {
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

	fullTimestamp := d.ThrownAt.Format("2006-01-02 03:04:05 PM MST")

	return Div(
		Class(containerClass+" mb-2"),
		Title(fullTimestamp),
		Div(
			Class("max-w-[70%]"),
			Div(
				Class("px-4 py-2 "+bubbleClass+" flex items-center gap-2"),
				Img(
					Src("/images/rock.svg"),
					Alt("Rock thrown"),
					Class("w-5 h-5 flex-shrink-0"),
				),
				g.Text("Rock thrown"),
			),
			Div(
				Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1 px-1"),
				g.Text(formatMessageTime(d.ThrownAt)),
			),
		),
	)
}

// MessageTimeline renders chat messages with an optional rock event inserted
// chronologically. Message and rock timestamps must already be in the viewer's
// timezone.
func MessageTimeline(messages []MessageItemData, rock *RockEventData) []g.Node {
	if rock == nil {
		nodes := make([]g.Node, len(messages))
		for i, m := range messages {
			nodes[i] = MessageItem(m)
		}
		return nodes
	}

	var nodes []g.Node
	inserted := false
	for _, m := range messages {
		if !inserted && m.CreatedAt.After(rock.ThrownAt) {
			nodes = append(nodes, RockThrownMessage(*rock))
			inserted = true
		}
		nodes = append(nodes, MessageItem(m))
	}
	if !inserted {
		nodes = append(nodes, RockThrownMessage(*rock))
	}
	return nodes
}

func ConversationMessagesSentinel(conversationID int) g.Node {
	return Div(
		ID(fmt.Sprintf("conversation-%d-sentinel", conversationID)),
		g.Attr("style", "display: none;"),
	)
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

func ConversationContentInput(conversationID int, attrs ...g.Node) g.Node {
	allAttrs := []g.Node{
		ID(fmt.Sprintf("conversation-%d-content-input", conversationID)),
		Type("text"),
		Name("content"),
		Placeholder("Type a message..."),
		Required(),
		Class("flex-1 px-4 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-200"),
		g.Attr("onkeydown", "if(event.key==='Enter' && !event.shiftKey) { event.preventDefault(); this.form.requestSubmit(); }"),
		g.Attr("hx-preserve", "true"), // Preserve input value during outerHTML swaps
	}
	allAttrs = append(allAttrs, attrs...)
	return Input(g.Group(allAttrs))
}

func ConversationForm(conversationID, adID int, csrfToken string, canPost bool,
	hasThrownRock, canThrowRock bool, hasPublicRock bool, messageCount int) g.Node {
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
			hx.Swap("none"),
		)
	}
	return Form(
		g.Group(attrs),
		Div(
			Class("flex gap-2"),
			g.If(canPost,
				ConversationContentInput(conversationID),
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
		// Rock throw link below the input (only shown if conversation has messages)
		g.If(messageCount > 0,
			Div(
				ID(fmt.Sprintf("%s-rock-link-container", modalName)),
				Class("mt-2"),
				RockThrowLink(conversationID, hasThrownRock, canThrowRock, csrfToken),
			),
		),
		g.If(messageCount == 0,
			Div(
				ID(fmt.Sprintf("%s-rock-link-container", modalName)),
				Class("mt-2"),
			),
		),
		g.If(canPost && hasPublicRock,
			Div(
				Class("mt-2"),
				RockOpinionLink(conversationID),
				RockOpinionIndicator(),
			),
		),
	)
}

// ConversationModalSwapOOB returns just the modal div (without backdrop) with hx-swap-oob="outerHTML" for updating via SSE or OOB swaps
// This is used to update the modal when messages are sent or rocks are thrown
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
			Div(
				ID(fmt.Sprintf("%s-messages", modalName)),
				Class("flex-1 overflow-y-auto p-4 space-y-2"),
				g.If(len(d.MessageNodes) == 0,
					Div(
						ID(fmt.Sprintf("conversation-%d-empty-message", d.ConversationID)),
						Class("text-center text-zinc-500 dark:text-zinc-400 py-8"),
						g.If(d.CanPost,
							g.Text("No messages yet. Start the conversation!"),
						),
						g.If(!d.CanPost,
							g.Text("No messages yet."),
						),
					),
				),
				g.If(len(d.MessageNodes) > 0,
					g.Group(d.MessageNodes),
				),
				ConversationMessagesSentinel(d.ConversationID),
			),
			ConversationForm(
				d.ConversationID, d.AdID, d.CSRFToken,
				d.CanPost, d.HasThrownRock, d.CanThrowRock,
				d.RockThrowerID != nil, len(d.MessageNodes),
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
				Div(
					ID(fmt.Sprintf("%s-messages", modalName)),
					Class("flex-1 overflow-y-auto p-4 space-y-2"),
					g.If(len(d.MessageNodes) == 0,
						Div(
							ID(fmt.Sprintf("conversation-%d-empty-message", d.ConversationID)),
							Class("text-center text-zinc-500 dark:text-zinc-400 py-8"),
							g.If(d.CanPost,
								g.Text("No messages yet. Start the conversation!"),
							),
							g.If(!d.CanPost,
								g.Text("No messages yet."),
							),
						),
					),
					g.If(len(d.MessageNodes) > 0,
						g.Group(d.MessageNodes),
					),
					ConversationMessagesSentinel(d.ConversationID),
				),
				ConversationForm(
					d.ConversationID, d.AdID, d.CSRFToken,
					d.CanPost, d.HasThrownRock, d.CanThrowRock,
					d.RockThrowerID != nil, len(d.MessageNodes),
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
		hx.Target("body"),
		hx.Swap("beforeend"),
		hx.Trigger("click[!closest(.rock-icon-container)]"),
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
					g.If(d.RockCount > 0, RockIcons(d.AdID, d.RockCount)),
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
							UserNameLink(d.OwnerID, d.OtherUserName, UserRockIcons(d.OwnerID, d.OtherUserRockCount)),
						}),
					),
					g.If(d.OwnerID == d.CurrentUserID,
						g.Group([]g.Node{
							g.Text("From: "),
							UserNameLink(d.InquirerID, d.OtherUserName, UserRockIcons(d.InquirerID, d.OtherUserRockCount)),
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
		Div(
			Class("mt-6"),
			Div(
				Class("bg-white dark:bg-zinc-800 rounded-lg shadow-lg border border-zinc-200 dark:border-zinc-700"),
				g.Group(conversationNodes),
			),
		),
	}
}

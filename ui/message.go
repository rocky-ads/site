package ui

import (
	"fmt"
	"time"

	"github.com/rocky-ads/site/message"
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

// RockThrownMessage renders a message-like item showing when a rock was thrown
func RockThrownMessage(throwerID, currentUserID int, thrownAt time.Time, loc *time.Location, ownerID, enquirerID int) g.Node {
	isSent := throwerID == currentUserID

	var bubbleClass string
	var containerClass string
	if isSent {
		bubbleClass = "bg-blue-500 text-white rounded-lg rounded-tr-none"
		containerClass = "flex justify-end"
	} else {
		bubbleClass = "bg-zinc-200 dark:bg-zinc-700 text-zinc-900 dark:text-zinc-200 rounded-lg rounded-tl-none"
		containerClass = "flex justify-start"
	}

	localTime := thrownAt.In(loc)
	fullTimestamp := localTime.Format("2006-01-02 03:04:05 PM MST")

	return Div(
		Class(containerClass+" mb-2"),
		Title(fullTimestamp),
		Div(
			Class("max-w-[70%]"),
			Div(
				Class("px-4 py-2 "+bubbleClass+" flex items-center gap-2"),
				Img(
					Src("/images/broken-egg.svg"),
					Alt("Rock thrown"),
					Class("w-5 h-5 flex-shrink-0"),
				),
				g.Text("Rock thrown"),
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

func MessageItemSwapOOB(conversationID int, messageNode g.Node) g.Node {
	sentinelID := fmt.Sprintf("conversation-%d-sentinel", conversationID)
	return Div(
		hx.SwapOOB(fmt.Sprintf("beforebegin:#%s", sentinelID)),
		messageNode,
	)
}

func ConversationEmptyMessageDeleteOOB(conversationID int, isFirstMessage bool) g.Node {
	if !isFirstMessage {
		return g.Raw("")
	}
	return Div(
		ID(fmt.Sprintf("conversation-%d-empty-message", conversationID)),
		hx.SwapOOB("delete"),
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

func ConversationListItemSwapOOB(conversationID int, conversationItemNode g.Node) g.Node {
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

func ConversationFormUpdateOOB(conversationID int, csrfToken string, hasThrownRock, canThrowRock bool) g.Node {
	modalName := fmt.Sprintf("conversation-%d", conversationID)
	return Form(
		ID(fmt.Sprintf("%s-form", modalName)),
		hx.SwapOOB("outerHTML"),
		Class("p-4 border-t border-zinc-200 dark:border-zinc-700 flex-shrink-0"),
		hx.Post(fmt.Sprintf("/auth/conversation/%d/send", conversationID)),
		hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
		hx.Include("this"),
		hx.Swap("none"),
		Div(
			Class("flex gap-2"),
			ConversationContentInput(conversationID),
			Button(
				Type("submit"),
				Class("px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors"),
				g.Text("Send"),
			),
		),
		// Rock throw link below the input (conversation now has messages)
		Div(
			ID(fmt.Sprintf("%s-rock-link-container", modalName)),
			Class("mt-2"),
			RockThrowLink(conversationID, hasThrownRock, canThrowRock, csrfToken),
		),
	)
}

func ConversationForm(conversationID, adID int, csrfToken string, canPost bool, hasThrownRock, canThrowRock bool, messageCount int) g.Node {
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
	)
}

// RockThrowLinkSwapOOB returns an OOB swap node to insert the rock throw link into the rock link container below the form
// oldModalID is the modal ID currently in the DOM (e.g., "conversation-0" for new conversations)
// conversationID is the actual conversation ID for the rock link functionality
func RockThrowLinkSwapOOB(oldModalID string, conversationID int, hasThrownRock, canThrowRock bool, csrfToken string) g.Node {
	rockLinkContainerID := fmt.Sprintf("%s-rock-link-container", oldModalID)
	rockLink := RockThrowLink(conversationID, hasThrownRock, canThrowRock, csrfToken)
	// Insert into the rock link container below the form
	return Div(
		ID(rockLinkContainerID),
		hx.SwapOOB("outerHTML"),
		Class("mt-2"),
		rockLink,
	)
}

func ConversationModal(conversationID, adID int, adTitle string, ownerID, enquirerID, currentUserID int, otherUserName string, messageNodes []g.Node, csrfToken string) g.Node {
	// Default conversation with no rock
	conv := message.Conversation{
		RockThrowerID: nil,
		RockThrownAt:  nil,
	}
	// Default to allowing posting (for backward compatibility)
	return ConversationModalWithRock(conversationID, adID, ownerID, enquirerID, currentUserID, 0, 0, adTitle, otherUserName, otherUserName, csrfToken, true, false, false, messageNodes, conv)
}

// ConversationModalSwapOOB returns just the modal div (without backdrop) with hx-swap-oob="outerHTML" for updating via SSE or OOB swaps
// This is used to update the modal when messages are sent or rocks are thrown
// targetModalID is optional - if provided, it's used as the swap target ID (for new conversations transitioning from conversation-0-modal)
func ConversationModalSwapOOB(conversationID, adID, ownerID, enquirerID, currentUserID, enquirerRockCount, ownerRockCount int, adTitle, ownerName, enquirerName, csrfToken string, canPost, hasThrownRock, canThrowRock bool, messageNodes []g.Node, conv message.Conversation, targetModalID ...string) g.Node {
	modalName := fmt.Sprintf("conversation-%d", conversationID)
	modalID := modalName + "-modal"
	// Use targetModalID if provided (for new conversations), otherwise use the conversation ID modal
	if len(targetModalID) > 0 && targetModalID[0] != "" {
		modalID = targetModalID[0]
	}

	var rockUserID int
	if conv.RockThrowerID != nil {
		rockUserID = *conv.RockThrowerID
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
						g.Text(adTitle),
					),
					Div(
						Class("text-sm text-zinc-600 dark:text-zinc-400"),
						Span(Class("font-semibold"), g.Text("From: ")),
						UserRockIcons(enquirerID, enquirerRockCount),
						g.Text(enquirerName),
						g.Text(", "),
						Span(Class("font-semibold"), g.Text("To: ")),
						UserRockIcons(ownerID, ownerRockCount),
						g.Text(ownerName),
						g.Text(" (ad owner)"),
						g.If(!canPost && (ownerID != currentUserID && enquirerID != currentUserID),
							Span(
								Class("ml-2 px-2 py-0.5 bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 rounded text-xs"),
								g.Text("Read-only"),
							),
						),
					),
					g.If(!canPost && conv.RockThrowerID != nil,
						Div(
							Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1"),
							g.Text("Rock thrown by: "),
							g.If(rockUserID == currentUserID,
								Span(Class("text-blue-600 dark:text-blue-400 font-medium"),
									g.If(rockUserID == ownerID, g.Text(ownerName)),
									g.If(rockUserID == enquirerID, g.Text(enquirerName)),
								),
							),
							g.If(rockUserID != currentUserID,
								Span(Class("text-zinc-700 dark:text-zinc-300 font-medium"),
									g.If(rockUserID == ownerID, g.Text(ownerName)),
									g.If(rockUserID == enquirerID, g.Text(enquirerName)),
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
				ConversationMessagesSentinel(conversationID),
			),
			ConversationForm(conversationID, adID, csrfToken, canPost, hasThrownRock, canThrowRock, len(messageNodes)),
		),
	)
}

func ConversationModalWithRock(conversationID, adID, ownerID, enquirerID, currentUserID, enquirerRockCount, ownerRockCount int, adTitle, ownerName, enquirerName, csrfToken string, canPost, hasThrownRock, canThrowRock bool, messageNodes []g.Node, conv message.Conversation) g.Node {
	modalName := fmt.Sprintf("conversation-%d", conversationID)

	var rockUserID int
	if conv.RockThrowerID != nil {
		rockUserID = *conv.RockThrowerID
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
							g.Text(adTitle),
						),
						Div(
							Class("text-sm text-zinc-600 dark:text-zinc-400"),
							Span(Class("font-semibold"), g.Text("From: ")),
							UserRockIcons(enquirerID, enquirerRockCount),
							g.Text(enquirerName),
							g.Text(", "),
							Span(Class("font-semibold"), g.Text("To: ")),
							UserRockIcons(ownerID, ownerRockCount),
							g.Text(ownerName),
							g.Text(" (ad owner)"),
							g.If(!canPost && (ownerID != currentUserID && enquirerID != currentUserID),
								Span(
									Class("ml-2 px-2 py-0.5 bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 rounded text-xs"),
									g.Text("Read-only"),
								),
							),
						),
						g.If(!canPost && conv.RockThrowerID != nil,
							Div(
								Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1"),
								g.Text("Rock thrown by: "),
								g.If(rockUserID == currentUserID,
									// Current user threw it - use blue color (like sent messages)
									Span(Class("text-blue-600 dark:text-blue-400 font-medium"),
										g.If(rockUserID == ownerID, g.Text(ownerName)),
										g.If(rockUserID == enquirerID, g.Text(enquirerName)),
									),
								),
								g.If(rockUserID != currentUserID,
									// Someone else threw it - use gray color (like received messages)
									Span(Class("text-zinc-700 dark:text-zinc-300 font-medium"),
										g.If(rockUserID == ownerID, g.Text(ownerName)),
										g.If(rockUserID == enquirerID, g.Text(enquirerName)),
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
					ConversationMessagesSentinel(conversationID),
				),
				ConversationForm(conversationID, adID, csrfToken, canPost, hasThrownRock, canThrowRock, len(messageNodes)),
			),
		),
	})
}

func ConversationListItem(conversationID, adID, ownerID, enquirerID, currentUserID int, adTitle, lastMessageContent, otherUserName string, lastMessageAt *time.Time, updatedAt time.Time, hasUnread bool, rockCount int, otherUserRockCount int) g.Node {
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
		ID(fmt.Sprintf("conversation-item-%d", conversationID)),
		Class("border-b border-zinc-200 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-800 cursor-pointer transition-colors"),
		hx.Get(fmt.Sprintf("/auth/conversation/%d", conversationID)),
		hx.Target("body"),
		hx.Swap("beforeend"),
		hx.Trigger("click[!closest(.rock-icon-container)]"),
		Div(
			Class("p-4"),
			Div(
				Class("flex items-start justify-between mb-2"),
				Div(
					Class("flex items-center gap-2 flex-1 min-w-0"),
					g.If(hasUnread,
						Div(
							Class("bg-green-500 rounded-full w-2 h-2 flex-shrink-0"),
						),
					),
					g.If(rockCount > 0, RockIcons(adID, rockCount)),
					Span(
						Class("text-lg font-semibold text-zinc-900 dark:text-zinc-200"),
						g.Text(adTitle),
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
					g.If(enquirerID == currentUserID,
						g.Group([]g.Node{
							g.Text("To: "),
							UserRockIcons(ownerID, otherUserRockCount),
							g.Text(otherUserName),
						}),
					),
					g.If(ownerID == currentUserID,
						g.Group([]g.Node{
							g.Text("From: "),
							UserRockIcons(enquirerID, otherUserRockCount),
							g.Text(otherUserName),
						}),
					),
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

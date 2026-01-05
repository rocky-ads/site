package handler

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/ui"
	"github.com/rocky-ads/site/user"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func buildMessageNodes(messages []message.Message, currentUserID int, loc *time.Location) []g.Node {
	var messageNodes []g.Node
	for _, msg := range messages {
		messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
	}
	return messageNodes
}

func sendMessageSSE(conv message.Conversation, senderID int, msg message.Message) {
	// Determine recipient user ID
	recipientID := conv.OwnerID
	if conv.OwnerID == senderID {
		recipientID = conv.EnquirerID
	}

	// Send message update
	sendMessageUpdate(conv.ID, msg, recipientID)

	// Send indicator update - recipient always has unread messages after receiving a new message
	sendUnreadIndicatorUpdate(recipientID, true)

	// Send conversation list item update - show green dot for unread conversation
	sendConversationListItemUpdate(conv, recipientID, true)
}

func sendMessageUpdate(conversationID int, msg message.Message, recipientID int) {
	// Render message from recipient's perspective
	recipientMessageNode := ui.MessageItem(msg.SenderID, recipientID, msg.Content, msg.CreatedAt, time.UTC)
	recipientMessageSwapOOB := ui.MessageItemSwapOOB(conversationID, recipientMessageNode)
	messageHTML, err := renderToString(recipientMessageSwapOOB)
	if err != nil {
		logger.Error("Failed to render message for SSE", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	// Send message update
	SendSSEEvent(recipientID, SSEEvent{
		Event: "",
		Data:  messageHTML,
	})
}

func sendUnreadIndicatorUpdate(userID int, hasUnread bool) {
	indicatorSwapOOB := ui.UnreadIndicatorSwapOOB(hasUnread)
	indicatorHTML, err := renderToString(indicatorSwapOOB)
	if err != nil {
		logger.Error("Failed to render unread indicator for SSE", "error", err, "userID", userID)
		return
	}

	// Send indicator update
	SendSSEEvent(userID, SSEEvent{
		Event: "",
		Data:  indicatorHTML,
	})
}

func sendMessageAndRenderUpdate(c *fiber.Ctx, conv message.Conversation, currentUserID int,
	loc *time.Location) error {

	content := c.FormValue("content")
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Message content cannot be empty")
	}

	// Check if this is the first message (before creating it)
	existingMessages, _ := message.GetConversationMessages(conv.ID, currentUserID, loc)
	isFirstMessage := len(existingMessages) == 0

	msg, err := message.CreateMessage(conv.ID, currentUserID, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send message")
	}

	newMessageNode := ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc)
	messageSwapOOB := ui.MessageItemSwapOOB(conv.ID, newMessageNode)
	clearedInputOOB := ui.ConversationContentInput(conv.ID, hx.SwapOOB("true"))

	// If this is the first message, delete the "No messages yet" text
	var nodes []g.Node
	nodes = append(nodes, messageSwapOOB, clearedInputOOB)
	if isFirstMessage {
		emptyMessageDelete := Div(
			ID(fmt.Sprintf("conversation-%d-empty-message", conv.ID)),
			hx.SwapOOB("delete"),
		)
		nodes = append(nodes, emptyMessageDelete)
	}

	// Send SSE event to recipient with the new message
	sendMessageSSE(conv, currentUserID, msg)

	return render(c, g.Group(nodes))
}

func getOtherUserName(conv message.Conversation, currentUserID int) (string, error) {
	var otherUserID int
	if conv.OwnerID == currentUserID {
		otherUserID = conv.EnquirerID
	} else {
		otherUserID = conv.OwnerID
	}

	otherUser, err := user.GetByID(otherUserID)
	if err != nil {
		return "", err
	}
	return otherUser.Name, nil
}

func sendConversationListItemUpdate(conv message.Conversation, currentUserID int, hasUnread bool) {
	// Get ad title
	a, err := ad.GetAd(currentUserID, conv.AdID, time.UTC)
	if err != nil {
		logger.Error("Failed to get ad for conversation list item update", "error", err, "conversationID", conv.ID, "adID", conv.AdID)
		return
	}

	// Get other user name
	otherUserName, err := getOtherUserName(conv, currentUserID)
	if err != nil {
		logger.Error("Failed to get other user name for conversation list item update", "error", err, "conversationID", conv.ID)
		return
	}

	// Get last message for preview
	messages, err := message.GetConversationMessages(conv.ID, currentUserID, time.UTC)
	var lastMessageContent string
	var lastMessageAt *time.Time
	if err == nil && len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		lastMessageContent = lastMsg.Content
		lastMessageAt = &lastMsg.CreatedAt
	}

	// Render conversation list item (always includes ID)
	conversationItem := ui.ConversationListItem(
		conv.ID,
		conv.AdID,
		conv.OwnerID,
		conv.EnquirerID,
		currentUserID,
		a.Title,
		lastMessageContent,
		otherUserName,
		lastMessageAt,
		conv.UpdatedAt,
		hasUnread,
	)
	conversationItemSwapOOB := ui.ConversationListItemSwapOOB(conv.ID, conversationItem)
	itemHTML, err := renderToString(conversationItemSwapOOB)
	if err != nil {
		logger.Error("Failed to render conversation list item for SSE", "error", err, "conversationID", conv.ID)
		return
	}

	// Send conversation list item update
	SendSSEEvent(currentUserID, SSEEvent{
		Event: "",
		Data:  itemHTML,
	})
}

func renderConversationModal(c *fiber.Ctx, conv message.Conversation, currentUserID int, loc *time.Location, csrfToken string) error {
	a, err := ad.GetAd(currentUserID, conv.AdID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	messages, err := message.GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
	}

	otherUserName, err := getOtherUserName(conv, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get other user")
	}

	messageNodes := buildMessageNodes(messages, currentUserID, loc)

	return render(c, ui.ConversationModal(conv.ID, conv.AdID, a.Title, conv.OwnerID, conv.EnquirerID, currentUserID, otherUserName, messageNodes, csrfToken))
}

func MessageModalHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	a, err := ad.GetAd(currentUserID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	if a.UserID == currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "You cannot message your own ad")
	}

	conv, err := message.GetOrCreateConversation(adID, a.UserID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	return renderConversationModal(c, conv, currentUserID, loc, csrfToken)
}

func SendMessageHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)

	a, err := ad.GetAd(currentUserID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	conv, err := message.GetOrCreateConversation(adID, a.UserID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	return sendMessageAndRenderUpdate(c, conv, currentUserID, loc)
}

func SendConversationMessageHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)

	conv, err := message.GetConversation(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	return sendMessageAndRenderUpdate(c, conv, currentUserID, loc)
}

func ConversationModalHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	conv, err := message.GetConversation(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	// Mark conversation as read when opened
	if err := message.MarkConversationAsRead(conversationID, currentUserID); err != nil {
		// Log error but don't fail the request
		logger.Error("Failed to mark conversation as read", "error", err, "conversationID", conversationID, "userID", currentUserID)
	} else {
		// Send conversation list item update - remove green dot since conversation is now read
		sendConversationListItemUpdate(conv, currentUserID, false)

		// Check if there are any remaining unread conversations
		hasUnread, err := message.GetHasUnread(currentUserID)
		if err == nil {
			// Update avatar indicator based on whether there are any remaining unread conversations
			sendUnreadIndicatorUpdate(currentUserID, hasUnread)
		}
	}

	return renderConversationModal(c, conv, currentUserID, loc, csrfToken)
}

func UserMessagesHandler(c *fiber.Ctx) error {
	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)

	conversations, err := message.GetUserConversations(currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load conversations")
	}

	userMap := make(map[int]string)
	for _, conv := range conversations {
		if _, ok := userMap[conv.OtherUserID]; !ok {
			u, err := user.GetByID(conv.OtherUserID)
			if err == nil {
				userMap[conv.OtherUserID] = u.Name
			}
		}
	}

	var conversationItems []g.Node
	for _, conv := range conversations {
		otherUserName := userMap[conv.OtherUserID]
		if otherUserName == "" {
			otherUserName = "Unknown User"
		}
		conversationItems = append(conversationItems, ui.ConversationListItem(
			conv.ID,
			conv.AdID,
			conv.OwnerID,
			conv.EnquirerID,
			currentUserID,
			conv.AdTitle,
			conv.LastMessageContent,
			otherUserName,
			conv.LastMessageAt,
			conv.UpdatedAt,
			conv.HasUnread,
		))
	}

	return renderPage(c, "Messages", ui.MessagesPage(conversationItems))
}

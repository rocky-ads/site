package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/ui"
	"github.com/rocky-ads/site/user"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func GetMessageCount(userID int) int {
	count, err := message.GetUnreadMessageCount(userID)
	if err != nil {
		return 0
	}
	return count
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

	messages, err := message.GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
	}

	owner, err := user.GetByID(a.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get owner")
	}

	var messageNodes []g.Node
	for _, msg := range messages {
		messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
	}

	return render(c, ui.ConversationModal(conv.ID, adID, a.Title, a.UserID, currentUserID, currentUserID, owner.Name, messageNodes, csrfToken))
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

	content := c.FormValue("content")
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Message content cannot be empty")
	}

	_, err = message.CreateMessage(conv.ID, currentUserID, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send message")
	}

	messages, err := message.GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
	}

	modalName := fmt.Sprintf("conversation-%d", conv.ID)
	var messageNodes []g.Node
	for _, msg := range messages {
		messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
	}

	return render(c, Div(
		ID(modalName+"-messages"),
		hx.SwapOOB("true"),
		Class("flex-1 overflow-y-auto p-4 space-y-2"),
		g.Group(messageNodes),
	))
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

	content := c.FormValue("content")
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Message content cannot be empty")
	}

	_, err = message.CreateMessage(conv.ID, currentUserID, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send message")
	}

	messages, err := message.GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
	}

	modalName := fmt.Sprintf("conversation-%d", conv.ID)
	var messageNodes []g.Node
	for _, msg := range messages {
		messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
	}

	return render(c, Div(
		ID(modalName+"-messages"),
		hx.SwapOOB("true"),
		Class("flex-1 overflow-y-auto p-4 space-y-2"),
		g.Group(messageNodes),
	))
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

	a, err := ad.GetAd(currentUserID, conv.AdID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	messages, err := message.GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
	}

	var otherUserID int
	var otherUserName string
	if conv.OwnerID == currentUserID {
		otherUserID = conv.EnquirerID
	} else {
		otherUserID = conv.OwnerID
	}

	otherUser, err := user.GetByID(otherUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get other user")
	}
	otherUserName = otherUser.Name

	var messageNodes []g.Node
	for _, msg := range messages {
		messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
	}

	return render(c, ui.ConversationModal(conv.ID, conv.AdID, a.Title, conv.OwnerID, conv.EnquirerID, currentUserID, otherUserName, messageNodes, csrfToken))
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

	return renderPage(c, "Messages", ui.MessagesPage(conversations, userMap))
}

package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/egg"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/service/sms"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type conversationModalRenderMode int

const (
	modalRenderInitial conversationModalRenderMode = iota
	modalRenderSwapOOB
)

func conversationModalData(
	conv message.Conversation,
	currentUserID, enquirerEggCount, ownerEggCount int,
	adTitle, ownerName, enquirerName, csrfToken string,
	canPost, hasThrownEgg, canThrowEgg bool,
	messageNodes []g.Node,
	targetModalID string,
) ui.ConversationModalData {
	return ui.ConversationModalData{
		ConversationID:   conv.ID,
		AdID:             conv.AdID,
		OwnerID:          conv.OwnerID,
		EnquirerID:       conv.EnquirerID,
		CurrentUserID:    currentUserID,
		EnquirerEggCount: enquirerEggCount,
		OwnerEggCount:    ownerEggCount,
		AdTitle:          adTitle,
		OwnerName:        ownerName,
		EnquirerName:     enquirerName,
		CSRFToken:        csrfToken,
		CanPost:          canPost,
		HasThrownEgg:     hasThrownEgg,
		CanThrowEgg:      canThrowEgg,
		MessageNodes:     messageNodes,
		EggThrowerID:     conv.EggThrowerID,
		TargetModalID:    targetModalID,
	}
}

func messageItemData(msg message.Message, currentUserID int, loc *time.Location) ui.MessageItemData {
	return ui.MessageItemData{
		SenderID:      msg.SenderID,
		CurrentUserID: currentUserID,
		Content:       msg.Content,
		CreatedAt:     msg.CreatedAt.In(loc),
	}
}

func eggEventData(throwerID, currentUserID, ownerID, enquirerID int, thrownAt time.Time, loc *time.Location) ui.EggEventData {
	return ui.EggEventData{
		ThrowerID:     throwerID,
		CurrentUserID: currentUserID,
		ThrownAt:      thrownAt.In(loc),
		OwnerID:       ownerID,
		EnquirerID:    enquirerID,
	}
}

func conversationListItemData(
	conversationID, adID, ownerID, enquirerID, currentUserID int,
	adTitle, lastMessageContent, otherUserName string,
	lastMessageAt *time.Time, updatedAt time.Time,
	hasUnread bool, eggCount, otherUserEggCount int,
	loc *time.Location,
) ui.ConversationListItemData {
	d := ui.ConversationListItemData{
		ConversationID:     conversationID,
		AdID:               adID,
		OwnerID:            ownerID,
		EnquirerID:         enquirerID,
		CurrentUserID:      currentUserID,
		AdTitle:            adTitle,
		LastMessageContent: lastMessageContent,
		OtherUserName:      otherUserName,
		UpdatedAt:          updatedAt.In(loc),
		HasUnread:          hasUnread,
		EggCount:           eggCount,
		OtherUserEggCount:  otherUserEggCount,
	}
	if lastMessageAt != nil {
		t := lastMessageAt.In(loc)
		d.LastMessageAt = &t
	}
	return d
}

func conversationListItemFromConv(conv message.ConversationWithLastMessage, currentUserID int) ui.ConversationListItemData {
	return ui.ConversationListItemData{
		ConversationID:     conv.ID,
		AdID:               conv.AdID,
		OwnerID:            conv.OwnerID,
		EnquirerID:         conv.EnquirerID,
		CurrentUserID:      currentUserID,
		AdTitle:            conv.AdTitle,
		LastMessageContent: conv.LastMessageContent,
		OtherUserName:      conv.OtherUserName,
		LastMessageAt:      conv.LastMessageAt,
		UpdatedAt:          conv.UpdatedAt,
		HasUnread:          conv.HasUnread,
		EggCount:           conv.RockCount,
		OtherUserEggCount:  conv.OtherUserEggCount,
	}
}

func messageTimelineFromView(view message.ConversationModalView, currentUserID int, loc *time.Location) []g.Node {
	msgs := make([]ui.MessageItemData, len(view.Messages))
	for i, msg := range view.Messages {
		msgs[i] = messageItemData(msg, currentUserID, loc)
	}
	var egg *ui.EggEventData
	conv := view.Conversation
	if conv.EggThrowerID != nil && conv.EggThrownAt != nil {
		e := eggEventData(
			*conv.EggThrowerID, currentUserID, conv.OwnerID, conv.EnquirerID,
			*conv.EggThrownAt, loc)
		egg = &e
	}
	return ui.MessageTimeline(msgs, egg)
}

func conversationModalDataFromView(
	view message.ConversationModalView,
	currentUserID int,
	csrfToken, targetModalID string,
	messageNodes []g.Node,
) ui.ConversationModalData {
	return conversationModalData(
		view.Conversation, currentUserID,
		view.EnquirerEggCount, view.OwnerEggCount,
		view.AdTitle, view.OwnerName, view.EnquirerName, csrfToken,
		view.CanPost, view.HasThrownEgg, view.CanThrowEgg,
		messageNodes, targetModalID,
	)
}

func buildConversationModalError(err error) error {
	if errors.Is(err, message.ErrModalAdNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	logger.Error("Failed to build conversation modal", "error", err)
	return fiber.NewError(fiber.StatusInternalServerError, "Failed to load conversation")
}

func renderConversationModalView(
	c *fiber.Ctx,
	view message.ConversationModalView,
	currentUserID int,
	loc *time.Location,
	csrfToken, targetModalID string,
	mode conversationModalRenderMode,
) error {
	messageNodes := messageTimelineFromView(view, currentUserID, loc)
	data := conversationModalDataFromView(view, currentUserID, csrfToken, targetModalID, messageNodes)
	if mode == modalRenderInitial {
		return render(c, ui.ConversationModalWithEgg(data))
	}
	return render(c, ui.ConversationModalSwapOOB(data))
}

func renderConversationModal(c *fiber.Ctx, conv message.Conversation, currentUserID int, loc *time.Location, csrfToken string) error {
	view, err := message.BuildConversationModal(conv, currentUserID, loc)
	if err != nil {
		return buildConversationModalError(err)
	}
	return renderConversationModalView(c, view, currentUserID, loc, csrfToken, "", modalRenderInitial)
}

func renderConversationModalSwapOOB(c *fiber.Ctx, conv message.Conversation, currentUserID int, loc *time.Location, csrfToken string) error {
	view, err := message.BuildConversationModal(conv, currentUserID, loc)
	if err != nil {
		return buildConversationModalError(err)
	}
	return renderConversationModalView(c, view, currentUserID, loc, csrfToken, "", modalRenderSwapOOB)
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
	conv, err := message.GetConversationByID(conversationID)
	if err != nil {
		logger.Error("Failed to get conversation for SSE modal update", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	loc := time.UTC
	view, err := message.BuildConversationModal(conv, recipientID, loc)
	if err != nil {
		logger.Error("Failed to build conversation modal for SSE", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	messageNodes := messageTimelineFromView(view, recipientID, loc)
	modalSwapOOB := ui.ConversationModalSwapOOB(conversationModalDataFromView(
		view, recipientID, "", "", messageNodes))
	modalHTML, err := renderToString(modalSwapOOB)
	if err != nil {
		logger.Error("Failed to render modal for SSE", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	SendSSEEvent(recipientID, SSEEvent{
		Event: "",
		Data:  modalHTML,
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
	loc *time.Location, isNewConversation bool) error {

	content := c.FormValue("content")
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Message content cannot be empty")
	}

	// Step 1: Update database - if it fails, stop
	msg, err := message.CreateMessage(conv.ID, currentUserID, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send message")
	}

	// Step 2: Send SSE updates
	sendMessageSSE(conv, currentUserID, msg)

	// Step 3: Enqueue SMS notification (non-blocking)
	// Determine recipient user ID
	recipientID := conv.OwnerID
	if conv.OwnerID == currentUserID {
		recipientID = conv.EnquirerID
	}
	if err := sms.EnqueueNotification(recipientID, conv.ID); err != nil {
		logger.Error("Failed to enqueue SMS notification",
			"error", err,
			"component", "message",
			"conversationID", conv.ID,
			"recipientID", recipientID,
			"senderID", currentUserID)
		// Don't fail the request, just log
	}

	// Get updated conversation to render full modal
	updatedConv, err := message.GetConversationByID(conv.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get updated conversation")
	}

	csrfToken := local.GetCSRFToken(c)
	view, err := message.BuildConversationModal(updatedConv, currentUserID, loc)
	if err != nil {
		return buildConversationModalError(err)
	}

	targetModalID := ""
	if isNewConversation {
		targetModalID = "conversation-0-modal"
	}

	return renderConversationModalView(c, view, currentUserID, loc, csrfToken, targetModalID, modalRenderSwapOOB)
}

func sendConversationListItemUpdate(conv message.Conversation, currentUserID int, hasUnread bool) {
	// Get ad title
	a, err := ad.GetAd(currentUserID, conv.AdID, time.UTC)
	if err != nil {
		logger.Error("Failed to get ad for conversation list item update", "error", err, "conversationID", conv.ID, "adID", conv.AdID)
		return
	}

	// Get other user name
	otherUserName, err := message.OtherUserName(conv, currentUserID)
	if err != nil {
		logger.Error("Failed to get other user name for conversation list item update", "error", err, "conversationID", conv.ID)
		return
	}

	// Get other user ID
	var otherUserID int
	if conv.OwnerID == currentUserID {
		otherUserID = conv.EnquirerID
	} else {
		otherUserID = conv.OwnerID
	}

	// Get egg count for the other user
	otherUserEggCount, _ := egg.GetEggCountForUser(otherUserID)

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
	conversationItem := ui.ConversationListItem(conversationListItemData(
		conv.ID, conv.AdID, conv.OwnerID, conv.EnquirerID, currentUserID,
		a.Title, lastMessageContent, otherUserName,
		lastMessageAt, conv.UpdatedAt,
		hasUnread, a.RockCount, otherUserEggCount,
		time.UTC,
	))
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

	// Try to get existing conversation
	conv, err := message.GetConversationByAdAndEnquirer(adID, a.UserID, currentUserID)
	if err != nil {
		if err == message.ErrConversationNotFound {
			// No conversation exists yet - create a temporary conversation struct for the modal
			// The conversation will be created when the first message is sent
			conv = message.Conversation{
				ID:           0, // 0 indicates conversation doesn't exist yet
				AdID:         adID,
				OwnerID:      a.UserID,
				EnquirerID:   currentUserID,
				EggThrowerID: nil,
				EggThrownAt:  nil,
			}
		} else {
			logger.Error("Failed to get conversation", "error", err, "adID", adID, "ownerID", a.UserID, "enquirerID", currentUserID)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
		}
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

	// Try to get existing conversation
	conv, err := message.GetConversationByAdAndEnquirer(adID, a.UserID, currentUserID)
	isNewConversation := false
	if err != nil {
		if err == message.ErrConversationNotFound {
			// Create conversation when sending first message
			conv, err = message.CreateConversation(adID, a.UserID, currentUserID)
			if err != nil {
				logger.Error("Failed to create conversation", "error", err, "adID", adID, "ownerID", a.UserID, "enquirerID", currentUserID)
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to create conversation")
			}
			isNewConversation = true
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
		}
	}

	return sendMessageAndRenderUpdate(c, conv, currentUserID, loc, isNewConversation)
}

func SendConversationMessageHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)

	// Check if user can post (must be participant)
	canPost, err := message.CanUserPost(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check permissions")
	}
	if !canPost {
		return fiber.NewError(fiber.StatusForbidden, "You can only read this conversation")
	}

	conv, err := message.GetConversation(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	return sendMessageAndRenderUpdate(c, conv, currentUserID, loc, false)
}

func ConversationModalHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	conv, markedRead, err := message.OpenConversation(conversationID, currentUserID)
	if errors.Is(err, message.ErrModalAccess) {
		return fiber.NewError(fiber.StatusForbidden, "Conversation not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	if markedRead {
		sendConversationListItemUpdate(conv, currentUserID, false)

		hasUnread, err := message.GetHasUnread(currentUserID)
		if err == nil {
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
		logger.Error("Failed to get user conversations", "error", err, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load conversations")
	}

	items := make([]ui.ConversationListItemData, len(conversations))
	for i, conv := range conversations {
		items[i] = conversationListItemFromConv(conv, currentUserID)
	}

	return renderPage(c, "Messages", ui.MessagesPage(items))
}

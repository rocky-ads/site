package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/param"
	"github.com/rocky-ads/site/internal/rock"
	"github.com/rocky-ads/site/internal/rockopinion"
	"github.com/rocky-ads/site/internal/service/sms"
	"github.com/rocky-ads/site/internal/ui"
	g "maragu.dev/gomponents"
)

type conversationModalRenderMode int

const (
	modalRenderInitial conversationModalRenderMode = iota
	modalRenderSwapOOB
	modalRenderMessagesSwapOOB
	modalRenderMessageActionsSwapOOB
)

func conversationModalData(conv message.Conversation, currentUserID,
	inquirerRockCount, ownerRockCount int, adTitle, ownerName, inquirerName,
	csrfToken string, canPost, hasThrownRock, canThrowRock bool,
	messageNodes []g.Node, targetModalID string) ui.ConversationModalData {
	return ui.ConversationModalData{
		ConversationID:    conv.ID,
		AdID:              conv.AdID,
		OwnerID:           conv.OwnerID,
		InquirerID:        conv.InquirerID,
		CurrentUserID:     currentUserID,
		InquirerRockCount: inquirerRockCount,
		OwnerRockCount:    ownerRockCount,
		AdTitle:           adTitle,
		OwnerName:         ownerName,
		InquirerName:      inquirerName,
		CSRFToken:         csrfToken,
		CanPost:           canPost,
		HasThrownRock:     hasThrownRock,
		CanThrowRock:      canThrowRock,
		MessageNodes:      messageNodes,
		RockThrowerID:     conv.RockThrowerID,
		TargetModalID:     targetModalID,
	}
}

func messageItemData(msg message.Message, currentUserID int,
	tz *time.Location) ui.MessageItemData {
	return ui.MessageItemData{
		SenderID:      msg.SenderID,
		CurrentUserID: currentUserID,
		Content:       msg.Content,
		CreatedAt:     msg.CreatedAt.In(tz),
	}
}

func rockEventData(event rock.Event, currentUserID, ownerID,
	inquirerID int) ui.RockEventData {
	kind := ui.RockEventThrown
	if event.Kind == rock.EventUnthrown {
		kind = ui.RockEventUnthrown
	}
	return ui.RockEventData{
		ThrowerID:     event.UserID,
		CurrentUserID: currentUserID,
		Kind:          kind,
		EventAt:       event.CreatedAt,
		OwnerID:       ownerID,
		InquirerID:    inquirerID,
	}
}

func rockEventsFromView(view message.ConversationModalView,
	currentUserID int) []ui.RockEventData {
	conv := view.Conversation
	events := make([]ui.RockEventData, len(view.RockEvents))
	for i, event := range view.RockEvents {
		events[i] = rockEventData(event, currentUserID, conv.OwnerID,
			conv.InquirerID)
	}
	return events
}

func messageTimelineFromView(view message.ConversationModalView,
	currentUserID int, tz *time.Location) []g.Node {
	msgs := make([]ui.MessageItemData, len(view.Messages))
	for i, msg := range view.Messages {
		msgs[i] = messageItemData(msg, currentUserID, tz)
	}
	return ui.MessageTimeline(msgs, rockEventsFromView(view, currentUserID))
}

func conversationListItemData(conversationID, adID, ownerID, inquirerID,
	currentUserID int, adTitle, lastMessageContent, otherUserName string,
	lastMessageAt *time.Time, updatedAt time.Time, hasUnread bool, rockCount,
	otherUserRockCount int, tz *time.Location) ui.ConversationListItemData {
	d := ui.ConversationListItemData{
		ConversationID:     conversationID,
		AdID:               adID,
		OwnerID:            ownerID,
		InquirerID:         inquirerID,
		CurrentUserID:      currentUserID,
		AdTitle:            adTitle,
		LastMessageContent: lastMessageContent,
		OtherUserName:      otherUserName,
		UpdatedAt:          updatedAt.In(tz),
		HasUnread:          hasUnread,
		RockCount:          rockCount,
		OtherUserRockCount: otherUserRockCount,
	}
	if lastMessageAt != nil {
		t := lastMessageAt.In(tz)
		d.LastMessageAt = &t
	}
	return d
}

func conversationListItemFromConv(conv message.ConversationWithLastMessage,
	currentUserID int) ui.ConversationListItemData {
	return ui.ConversationListItemData{
		ConversationID:     conv.ID,
		AdID:               conv.AdID,
		OwnerID:            conv.OwnerID,
		InquirerID:         conv.InquirerID,
		CurrentUserID:      currentUserID,
		AdTitle:            conv.AdTitle,
		LastMessageContent: conv.LastMessageContent,
		OtherUserName:      conv.OtherUserName,
		LastMessageAt:      conv.LastMessageAt,
		UpdatedAt:          conv.UpdatedAt,
		HasUnread:          conv.HasUnread,
		RockCount:          conv.RockCount,
		OtherUserRockCount: conv.OtherUserRockCount,
	}
}

func conversationModalDataFromView(view message.ConversationModalView,
	currentUserID int, csrfToken, targetModalID string, messageNodes []g.Node) ui.ConversationModalData {
	return conversationModalData(
		view.Conversation, currentUserID,
		view.InquirerRockCount, view.OwnerRockCount,
		view.AdTitle, view.OwnerName, view.InquirerName, csrfToken,
		view.CanPost, view.HasThrownRock, view.CanThrowRock,
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

func renderConversationModalView(c *fiber.Ctx,
	view message.ConversationModalView, currentUserID int, tz *time.Location,
	csrfToken, targetModalID string, mode conversationModalRenderMode) error {
	messageNodes := messageTimelineFromView(view, currentUserID, tz)
	data := conversationModalDataFromView(view, currentUserID, csrfToken, targetModalID, messageNodes)
	switch mode {
	case modalRenderInitial:
		return render(c, ui.ConversationModalWithRock(data))
	case modalRenderSwapOOB:
		return render(c, ui.ConversationModalSwapOOB(data))
	case modalRenderMessagesSwapOOB:
		return render(c, g.Group([]g.Node{
			ui.ConversationMessagesSwapOOB(data),
			ui.ConversationContentInputClearSwapOOB(data.ConversationID),
		}))
	case modalRenderMessageActionsSwapOOB:
		return render(c, g.Group([]g.Node{
			ui.ConversationMessageActionsSwapOOB(data),
			ui.ConversationMessagesSwapOOB(data),
		}))
	default:
		return render(c, ui.ConversationModalSwapOOB(data))
	}
}

func renderConversationModal(c *fiber.Ctx, conv message.Conversation,
	currentUserID int, tz *time.Location, csrfToken string) error {
	view, err := message.BuildConversationModal(conv, currentUserID, tz)
	if err != nil {
		return buildConversationModalError(err)
	}
	return renderConversationModalView(c, view, currentUserID, tz, csrfToken, "", modalRenderInitial)
}

func renderConversationMessageActionsSwapOOB(c *fiber.Ctx, conv message.Conversation,
	currentUserID int, tz *time.Location, csrfToken string) error {
	view, err := message.BuildConversationModal(conv, currentUserID, tz)
	if err != nil {
		return buildConversationModalError(err)
	}
	return renderConversationModalView(c, view, currentUserID, tz, csrfToken, "", modalRenderMessageActionsSwapOOB)
}

func sendMessageSSE(conv message.Conversation, senderID int,
	msg message.Message) {
	// Determine recipient user ID
	recipientID := conv.OwnerID
	if conv.OwnerID == senderID {
		recipientID = conv.InquirerID
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

	tz := time.UTC
	view, err := message.BuildConversationModal(conv, recipientID, tz)
	if err != nil {
		logger.Error("Failed to build conversation modal for SSE", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	messageNodes := messageTimelineFromView(view, recipientID, tz)
	messagesSwapOOB := ui.ConversationMessagesSwapOOB(conversationModalDataFromView(
		view, recipientID, "", "", messageNodes))
	modalHTML, err := renderToString(messagesSwapOOB)
	if err != nil {
		logger.Error("Failed to render messages for SSE", "error", err, "conversationID", conversationID, "recipientID", recipientID)
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

func sendMessageAndRenderUpdate(c *fiber.Ctx, conv message.Conversation,
	currentUserID int, tz *time.Location, isNewConversation bool) error {

	content := c.FormValue("content")
	if content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Message content cannot be empty")
	}

	// Step 1: Update database - if it fails, stop
	msg, err := message.CreateMessage(conv.ID, currentUserID, content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send message")
	}
	if err := rockopinion.Invalidate(conv.ID); err != nil {
		logger.Error("Failed to invalidate rock opinion",
			"error", err, "conversationID", conv.ID)
	}

	// Step 2: Send SSE updates
	sendMessageSSE(conv, currentUserID, msg)

	// Step 3: Enqueue SMS notification (non-blocking)
	// Determine recipient user ID
	recipientID := conv.OwnerID
	if conv.OwnerID == currentUserID {
		recipientID = conv.InquirerID
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
	view, err := message.BuildConversationModal(updatedConv, currentUserID, tz)
	if err != nil {
		return buildConversationModalError(err)
	}

	targetModalID := ""
	renderMode := modalRenderMessagesSwapOOB
	if isNewConversation {
		targetModalID = "conversation-0-modal"
		renderMode = modalRenderSwapOOB
	}

	return renderConversationModalView(c, view, currentUserID, tz, csrfToken, targetModalID, renderMode)
}

func sendConversationListItemUpdate(conv message.Conversation, currentUserID int,
	hasUnread bool) {
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
		otherUserID = conv.InquirerID
	} else {
		otherUserID = conv.OwnerID
	}

	// Get rock count for the other user
	otherUserRockCount, _ := rock.GetRockCountForUser(otherUserID)

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
		conv.ID, conv.AdID, conv.OwnerID, conv.InquirerID, currentUserID,
		a.Title, lastMessageContent, otherUserName,
		lastMessageAt, conv.UpdatedAt,
		hasUnread, a.RockCount, otherUserRockCount,
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
	tz := cookie.GetTimezone(c)
	csrfToken := local.GetCSRFToken(c)

	a, err := ad.GetAd(currentUserID, adID, tz)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	if a.UserID == currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "You cannot message your own ad")
	}

	// Try to get existing conversation
	conv, err := message.GetConversationByAdAndInquirer(adID, a.UserID, currentUserID)
	if err != nil {
		if err == message.ErrConversationNotFound {
			// No conversation exists yet - create a temporary conversation struct for the modal
			// The conversation will be created when the first message is sent
			conv = message.Conversation{
				ID:            0, // 0 indicates conversation doesn't exist yet
				AdID:          adID,
				OwnerID:       a.UserID,
				InquirerID:    currentUserID,
				RockThrowerID: nil,
				RockThrownAt:  nil,
			}
		} else {
			logger.Error("Failed to get conversation", "error", err, "adID", adID, "ownerID", a.UserID, "inquirerID", currentUserID)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
		}
	}

	return renderConversationModal(c, conv, currentUserID, tz, csrfToken)
}

func SendMessageHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	currentUserID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)

	a, err := ad.GetAd(currentUserID, adID, tz)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	// Try to get existing conversation
	conv, err := message.GetConversationByAdAndInquirer(adID, a.UserID, currentUserID)
	isNewConversation := false
	if err != nil {
		if err == message.ErrConversationNotFound {
			// Create conversation when sending first message
			conv, err = message.CreateConversation(adID, a.UserID, currentUserID)
			if err != nil {
				logger.Error("Failed to create conversation", "error", err, "adID", adID, "ownerID", a.UserID, "inquirerID", currentUserID)
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to create conversation")
			}
			isNewConversation = true
		} else {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
		}
	}

	return sendMessageAndRenderUpdate(c, conv, currentUserID, tz, isNewConversation)
}

func SendConversationMessageHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)

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

	return sendMessageAndRenderUpdate(c, conv, currentUserID, tz, false)
}

func rockOpinionModalData(conv message.Conversation, currentUserID int,
	tz *time.Location) (ui.RockOpinionModalData, error) {
	ownerName, inquirerName, err := message.OwnerAndInquirerNames(conv)
	if err != nil {
		return ui.RockOpinionModalData{}, err
	}

	a, err := ad.GetAd(currentUserID, conv.AdID, tz)
	if err != nil {
		return ui.RockOpinionModalData{}, err
	}

	d := ui.RockOpinionModalData{
		ConversationID: conv.ID,
		AdID:           conv.AdID,
		AdTitle:        a.Title,
		OwnerID:        conv.OwnerID,
		InquirerID:     conv.InquirerID,
		OwnerName:      ownerName,
		InquirerName:   inquirerName,
		CurrentUserID:  currentUserID,
		AdFacts:        rockopinion.AdFactLines(a, tz),
	}
	if conv.RockThrowerID != nil {
		d.RockThrowerID = *conv.RockThrowerID
	}

	op, err := rockopinion.GetOrGenerate(conv, tz)
	if errors.Is(err, rockopinion.ErrUnavailable) {
		d.Unavailable = true
		return d, nil
	}
	if err != nil {
		return ui.RockOpinionModalData{}, err
	}

	d.Summary = op.Summary
	d.AssessmentScore = op.Assessment
	d.AssessmentDetail = op.AssessmentDetail
	d.Resolution = op.Resolution
	d.Reasoning = op.Reasoning
	genAt := op.GeneratedAt
	if tz != nil {
		genAt = genAt.In(tz)
	}
	d.GeneratedAt = &genAt
	return d, nil
}

func renderRockOpinionModal(c *fiber.Ctx, conv message.Conversation,
	currentUserID int, tz *time.Location) error {
	d, err := rockOpinionModalData(conv, currentUserID, tz)
	if err != nil {
		logger.Error("Failed to build rock opinion modal", "error", err)
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"Failed to load dispute assessment",
		)
	}
	return render(c, ui.RockOpinionModal(d))
}

func RockOpinionHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)

	conv, err := message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}
	if !message.CanViewConversation(conv, currentUserID) {
		return fiber.NewError(fiber.StatusForbidden, "Conversation not found")
	}
	if conv.RockThrowerID == nil {
		return fiber.NewError(fiber.StatusNotFound, "No dispute assessment")
	}

	return renderRockOpinionModal(c, conv, currentUserID, tz)
}

func ConversationModalHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)
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

	return renderConversationModal(c, conv, currentUserID, tz, csrfToken)
}

func UserMessagesHandler(c *fiber.Ctx) error {
	currentUserID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)

	conversations, err := message.GetUserConversations(currentUserID, tz)
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

package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/rock"
	"github.com/rocky-ads/site/service/sms"
	"github.com/rocky-ads/site/ui"
	"github.com/rocky-ads/site/user"
	g "maragu.dev/gomponents"
)

func buildMessageNodes(messages []message.Message, currentUserID int, loc *time.Location) []g.Node {
	var messageNodes []g.Node
	for _, msg := range messages {
		messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
	}
	return messageNodes
}

func buildMessageNodesWithRock(messages []message.Message, currentUserID int, loc *time.Location, conv message.Conversation, ownerID, enquirerID int) []g.Node {
	var messageNodes []g.Node

	// Insert rock-thrown message if it exists
	if conv.RockThrowerID != nil && conv.RockThrownAt != nil {
		// Find the right position to insert the rock message (chronologically)
		inserted := false
		for _, msg := range messages {
			if !inserted && msg.CreatedAt.After(*conv.RockThrownAt) {
				// Insert rock message before this message
				rockThrownNode := ui.RockThrownMessage(*conv.RockThrowerID, currentUserID, *conv.RockThrownAt, loc, ownerID, enquirerID)
				messageNodes = append(messageNodes, rockThrownNode)
				inserted = true
			}
			messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
		}
		// If rock was thrown after all messages, append it at the end
		if !inserted {
			rockThrownNode := ui.RockThrownMessage(*conv.RockThrowerID, currentUserID, *conv.RockThrownAt, loc, ownerID, enquirerID)
			messageNodes = append(messageNodes, rockThrownNode)
		}
	} else {
		// No rock, just add messages normally
		for _, msg := range messages {
			messageNodes = append(messageNodes, ui.MessageItem(msg.SenderID, currentUserID, msg.Content, msg.CreatedAt, loc))
		}
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
	// Get the conversation to render the full modal
	conv, err := message.GetConversationByID(conversationID)
	if err != nil {
		logger.Error("Failed to get conversation for SSE modal update", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	// Render the entire modal with updated messages
	// Use UTC for SSE updates (no user-specific timezone available)
	loc := time.UTC
	messages, err := message.GetConversationMessages(conversationID, recipientID, loc)
	if err != nil {
		logger.Error("Failed to get messages for SSE modal update", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	// Build message nodes
	messageNodes := buildMessageNodesWithRock(messages, recipientID, loc, conv, conv.OwnerID, conv.EnquirerID)

	// Get ad info
	a, err := ad.GetAd(recipientID, conv.AdID, loc)
	if err != nil {
		logger.Error("Failed to get ad for SSE modal update", "error", err, "conversationID", conversationID, "adID", conv.AdID)
		return
	}

	// Get owner and enquirer names
	ownerName, enquirerName, err := getOwnerAndEnquirerNames(conv)
	if err != nil {
		logger.Error("Failed to get user names for SSE modal update", "error", err, "conversationID", conversationID)
		return
	}

	// Get rock counts
	enquirerRockCount, _ := rock.GetRockCountForUser(conv.EnquirerID)
	ownerRockCount, _ := rock.GetRockCountForUser(conv.OwnerID)

	// Check permissions
	canPost, err := message.CanUserPost(conversationID, recipientID)
	if err != nil {
		logger.Error("Failed to check permissions for SSE modal update", "error", err, "conversationID", conversationID)
		canPost = false
	}

	hasThrownRock, _ := rock.HasUserThrownRock(recipientID, conversationID)
	rockCount, _ := rock.GetRockCountForConversation(conversationID)
	userRockCount, _ := rock.GetUserRockCount(recipientID)
	canThrowRock := canPost && rockCount == 0 && userRockCount < 3

	// Render modal div with OOB swap (no CSRF token needed for SSE - it's read-only)
	modalSwapOOB := ui.ConversationModalSwapOOB(
		conversationID,
		conv.AdID,
		conv.OwnerID,
		conv.EnquirerID,
		recipientID,
		enquirerRockCount,
		ownerRockCount,
		a.Title,
		ownerName,
		enquirerName,
		"", // No CSRF token for SSE updates
		canPost,
		hasThrownRock,
		canThrowRock,
		messageNodes,
		conv,
	)
	modalHTML, err := renderToString(modalSwapOOB)
	if err != nil {
		logger.Error("Failed to render modal for SSE", "error", err, "conversationID", conversationID, "recipientID", recipientID)
		return
	}

	// Send modal update
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

	// Get all data needed for modal
	messages, err := message.GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
	}

	messageNodes := buildMessageNodesWithRock(messages, currentUserID, loc, updatedConv, updatedConv.OwnerID, updatedConv.EnquirerID)

	a, err := ad.GetAd(currentUserID, updatedConv.AdID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	ownerName, enquirerName, err := getOwnerAndEnquirerNames(updatedConv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get user names")
	}

	enquirerRockCount, _ := rock.GetRockCountForUser(updatedConv.EnquirerID)
	ownerRockCount, _ := rock.GetRockCountForUser(updatedConv.OwnerID)

	canPost, err := message.CanUserPost(conv.ID, currentUserID)
	if err != nil {
		canPost = false
	}

	hasThrownRock, _ := rock.HasUserThrownRock(currentUserID, conv.ID)
	rockCount, _ := rock.GetRockCountForConversation(conv.ID)
	userRockCount, _ := rock.GetUserRockCount(currentUserID)
	canThrowRock := canPost && rockCount == 0 && userRockCount < 3

	// Determine target modal ID for OOB swap
	// For new conversations, target conversation-0-modal; for existing, target conversation-{id}-modal
	targetModalID := ""
	if isNewConversation {
		targetModalID = "conversation-0-modal"
	}

	// Render modal div with OOB swap
	modalSwapOOB := ui.ConversationModalSwapOOB(
		conv.ID,
		updatedConv.AdID,
		updatedConv.OwnerID,
		updatedConv.EnquirerID,
		currentUserID,
		enquirerRockCount,
		ownerRockCount,
		a.Title,
		ownerName,
		enquirerName,
		csrfToken,
		canPost,
		hasThrownRock,
		canThrowRock,
		messageNodes,
		updatedConv,
		targetModalID,
	)

	return render(c, modalSwapOOB)
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

func getOwnerAndEnquirerNames(conv message.Conversation) (ownerName, enquirerName string, err error) {
	owner, err := user.GetByID(conv.OwnerID)
	if err != nil {
		return "", "", err
	}
	enquirer, err := user.GetByID(conv.EnquirerID)
	if err != nil {
		return "", "", err
	}
	return owner.Name, enquirer.Name, nil
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

	// Get other user ID
	var otherUserID int
	if conv.OwnerID == currentUserID {
		otherUserID = conv.EnquirerID
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
		a.RockCount,
		otherUserRockCount,
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

	var messages []message.Message
	var messageNodes []g.Node
	var canPost bool
	var hasThrownRock bool
	var canThrowRock bool

	// Handle case where conversation doesn't exist yet (ID == 0)
	if conv.ID == 0 {
		// New conversation - no messages yet, user can post first message
		messages = []message.Message{}
		messageNodes = []g.Node{}
		canPost = true
		hasThrownRock = false
		canThrowRock = false // Can't throw rock until conversation exists
	} else {
		// Existing conversation
		messages, err = message.GetConversationMessages(conv.ID, currentUserID, loc)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
		}

		messageNodes = buildMessageNodesWithRock(messages, currentUserID, loc, conv, conv.OwnerID, conv.EnquirerID)

		// Check if user can post (must be participant)
		canPost, err = message.CanUserPost(conv.ID, currentUserID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to check permissions")
		}

		// Check if user has thrown a rock
		hasThrownRock, err = rock.HasUserThrownRock(currentUserID, conv.ID)
		if err != nil {
			logger.Error("Failed to check if user threw rock", "error", err, "conversationID", conv.ID, "userID", currentUserID)
			hasThrownRock = false
		}

		// Check if ANY rock exists on this conversation
		rockCount, err := rock.GetRockCountForConversation(conv.ID)
		if err != nil {
			logger.Error("Failed to get rock count for conversation", "error", err, "conversationID", conv.ID)
			rockCount = 0
		}
		hasAnyRock := rockCount > 0

		// Check if user can throw rock (must be participant, have < 3 rocks, AND no rock exists on this conversation)
		canThrowRock = false
		if canPost && !hasAnyRock {
			userRockCount, err := rock.GetUserRockCount(currentUserID)
			if err == nil && userRockCount < 3 {
				canThrowRock = true
			}
		}
	}

	// Get owner and enquirer names
	ownerName, enquirerName, err := getOwnerAndEnquirerNames(conv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get user names")
	}

	// Get rock counts for owner and enquirer
	enquirerRockCount, _ := rock.GetRockCountForUser(conv.EnquirerID)
	ownerRockCount, _ := rock.GetRockCountForUser(conv.OwnerID)

	return render(c, ui.ConversationModalWithRock(
		conv.ID,
		conv.AdID,
		conv.OwnerID,
		conv.EnquirerID,
		currentUserID,
		enquirerRockCount,
		ownerRockCount,
		a.Title,
		ownerName,
		enquirerName,
		csrfToken,
		canPost,
		hasThrownRock,
		canThrowRock,
		messageNodes,
		conv,
	))
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
				ID:            0, // 0 indicates conversation doesn't exist yet
				AdID:          adID,
				OwnerID:       a.UserID,
				EnquirerID:    currentUserID,
				RockThrowerID: nil,
				RockThrownAt:  nil,
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

	// Get conversation (public conversations are accessible to all logged-in users)
	conv, err := message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	// Check if conversation is public or user is participant
	if conv.RockThrowerID == nil && conv.OwnerID != currentUserID && conv.EnquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Conversation not found")
	}

	// Mark conversation as read when opened (only if user is participant)
	if conv.OwnerID == currentUserID || conv.EnquirerID == currentUserID {
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

		// Get ad to get rock count for rock icons
		a, err := ad.GetAd(currentUserID, conv.AdID, loc)
		rockCount := 0
		if err == nil {
			rockCount = a.RockCount
		}

		// Get rock count for the other user
		otherUserRockCount, _ := rock.GetRockCountForUser(conv.OtherUserID)

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
			rockCount,
			otherUserRockCount,
		))
	}

	return renderPage(c, "Messages", ui.MessagesPage(conversationItems))
}

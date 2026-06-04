package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/egg"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/ui"
)

func ThrowEggHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	// Verify conversation exists and user has access
	conv, err := message.GetConversation(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	// Check if user is participant (only participants can throw eggs)
	if conv.OwnerID != currentUserID && conv.EnquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can throw eggs")
	}

	// Throw the egg
	err = egg.ThrowEgg(currentUserID, conversationID)
	if err != nil {
		if err == egg.ErrMaxEggsReached {
			return fiber.NewError(fiber.StatusBadRequest, "You have reached the maximum of 3 outstanding eggs")
		}
		if err == egg.ErrEggAlreadyThrown {
			return fiber.NewError(fiber.StatusBadRequest, "An egg has already been thrown at this conversation")
		}
		logger.Error("Failed to throw egg", "error", err, "conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to throw egg")
	}

	// Re-fetch conversation to get updated state
	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	// Render updated conversation modal
	return renderConversationModalWithEgg(c, conv, currentUserID, loc, csrfToken)
}

func UnthrowEggHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	// Verify conversation exists and user has access
	conv, err := message.GetConversation(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	// Check if user is participant (only participants can unthrow eggs)
	if conv.OwnerID != currentUserID && conv.EnquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can remove eggs")
	}

	// Unthrow the egg
	err = egg.UnthrowEgg(currentUserID, conversationID)
	if err != nil {
		if err == egg.ErrEggNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Egg not found")
		}
		logger.Error("Failed to unthrow egg", "error", err, "conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove egg")
	}

	// Re-fetch conversation to get updated state
	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	// Render updated conversation modal
	return renderConversationModalWithEgg(c, conv, currentUserID, loc, csrfToken)
}

func renderConversationModalWithEgg(c *fiber.Ctx, conv message.Conversation, currentUserID int, loc *time.Location, csrfToken string) error {
	a, err := ad.GetAd(currentUserID, conv.AdID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	messages, err := message.GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get messages")
	}

	// Get owner and enquirer names
	ownerName, enquirerName, err := getOwnerAndEnquirerNames(conv)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get user names")
	}

	// Get egg counts for owner and enquirer
	enquirerEggCount, _ := egg.GetEggCountForUser(conv.EnquirerID)
	ownerEggCount, _ := egg.GetEggCountForUser(conv.OwnerID)

	messageNodes := buildMessageNodesWithEgg(messages, currentUserID, loc, conv, conv.OwnerID, conv.EnquirerID)

	// Check if user can post (must be participant)
	canPost, err := message.CanUserPost(conv.ID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check permissions")
	}

	// Check if user has thrown an egg
	hasThrownEgg, err := egg.HasUserThrownEgg(currentUserID, conv.ID)
	if err != nil {
		logger.Error("Failed to check if user threw egg", "error", err, "conversationID", conv.ID, "userID", currentUserID)
		hasThrownEgg = false
	}

	// Check if ANY egg exists on this conversation
	eggCount, err := egg.GetEggCountForConversation(conv.ID)
	if err != nil {
		logger.Error("Failed to get egg count for conversation", "error", err, "conversationID", conv.ID)
		eggCount = 0
	}
	hasAnyEgg := eggCount > 0

	// Check if user can throw egg (must be participant, have < 3 eggs, AND no egg exists on this conversation)
	canThrowEgg := false
	if canPost && !hasAnyEgg {
		userEggCount, err := egg.GetUserEggCount(currentUserID)
		if err == nil && userEggCount < 3 {
			canThrowEgg = true
		}
	}

	// Return modal with OOB swap to update the existing modal
	modalSwapOOB := ui.ConversationModalSwapOOB(conversationModalData(
		conv, currentUserID, enquirerEggCount, ownerEggCount,
		a.Title, ownerName, enquirerName, csrfToken,
		canPost, hasThrownEgg, canThrowEgg, messageNodes, "",
	))

	return render(c, modalSwapOOB)
}

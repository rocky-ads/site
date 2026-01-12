package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/rock"
	"github.com/rocky-ads/site/ui"
)

func ThrowRockHandler(c *fiber.Ctx) error {
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

	// Check if user is participant (only participants can throw rocks)
	if conv.OwnerID != currentUserID && conv.EnquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can throw rocks")
	}

	// Throw the rock
	err = rock.ThrowRock(currentUserID, conversationID)
	if err != nil {
		if err == rock.ErrMaxRocksReached {
			return fiber.NewError(fiber.StatusBadRequest, "You have reached the maximum of 3 outstanding rocks")
		}
		if err == rock.ErrRockAlreadyThrown {
			return fiber.NewError(fiber.StatusBadRequest, "A rock has already been thrown at this conversation")
		}
		logger.Error("Failed to throw rock", "error", err, "conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to throw rock")
	}

	// Re-fetch conversation to get updated state
	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	// Render updated conversation modal
	return renderConversationModalWithRock(c, conv, currentUserID, loc, csrfToken)
}

func UnthrowRockHandler(c *fiber.Ctx) error {
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

	// Check if user is participant (only participants can unthrow rocks)
	if conv.OwnerID != currentUserID && conv.EnquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can remove rocks")
	}

	// Unthrow the rock
	err = rock.UnthrowRock(currentUserID, conversationID)
	if err != nil {
		if err == rock.ErrRockNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Rock not found")
		}
		logger.Error("Failed to unthrow rock", "error", err, "conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove rock")
	}

	// Re-fetch conversation to get updated state
	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	// Render updated conversation modal
	return renderConversationModalWithRock(c, conv, currentUserID, loc, csrfToken)
}

func renderConversationModalWithRock(c *fiber.Ctx, conv message.Conversation, currentUserID int, loc *time.Location, csrfToken string) error {
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

	// Get rock counts for owner and enquirer
	enquirerRockCount, _ := rock.GetRockCountForUser(conv.EnquirerID)
	ownerRockCount, _ := rock.GetRockCountForUser(conv.OwnerID)

	messageNodes := buildMessageNodesWithRock(messages, currentUserID, loc, conv, conv.OwnerID, conv.EnquirerID)

	// Check if user can post (must be participant)
	canPost, err := message.CanUserPost(conv.ID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check permissions")
	}

	// Check if user has thrown a rock
	hasThrownRock, err := rock.HasUserThrownRock(currentUserID, conv.ID)
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
	canThrowRock := false
	if canPost && !hasAnyRock {
		userRockCount, err := rock.GetUserRockCount(currentUserID)
		if err == nil && userRockCount < 3 {
			canThrowRock = true
		}
	}

	// Return modal with OOB swap to update the existing modal
	modalSwapOOB := ui.ConversationModalSwapOOB(
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
	)

	return render(c, modalSwapOOB)
}

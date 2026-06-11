package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/egg"
	"github.com/rocky-ads/site/internal/eggopinion"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
)

func ThrowEggHandler(c *fiber.Ctx) error {
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

	if conv.OwnerID != currentUserID && conv.InquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can throw eggs")
	}

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

	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	return renderConversationModalSwapOOB(c, conv, currentUserID, loc, csrfToken)
}

func UnthrowEggHandler(c *fiber.Ctx) error {
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

	if conv.OwnerID != currentUserID && conv.InquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can remove eggs")
	}

	err = egg.UnthrowEgg(currentUserID, conversationID)
	if err != nil {
		if err == egg.ErrEggNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Egg not found")
		}
		logger.Error("Failed to unthrow egg", "error", err, "conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove egg")
	}
	if err := eggopinion.Invalidate(conversationID); err != nil {
		logger.Error("Failed to invalidate egg opinion",
			"error", err, "conversationID", conversationID)
	}

	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	return renderConversationModalSwapOOB(c, conv, currentUserID, loc, csrfToken)
}

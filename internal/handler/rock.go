package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/rock"
	"github.com/rocky-ads/site/internal/rockopinion"
)

func ThrowRockHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)
	csrfToken := local.GetCSRFToken(c)

	conv, err := message.GetConversation(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	if conv.OwnerID != currentUserID && conv.InquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can throw rocks")
	}

	err = rock.ThrowRock(currentUserID, conversationID)
	if err != nil {
		if err == rock.ErrMaxRocksReached {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf(
				"You have reached the maximum of %d outstanding rocks",
				config.MaxOutstandingRocks,
			))
		}
		if err == rock.ErrRockAlreadyThrown {
			return fiber.NewError(fiber.StatusBadRequest, "An rock has already been thrown at this conversation")
		}
		logger.Error("Failed to throw rock", "error", err, "conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to throw rock")
	}

	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	return renderConversationModalSwapOOB(c, conv, currentUserID, tz, csrfToken)
}

func UnthrowRockHandler(c *fiber.Ctx) error {
	conversationID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid conversation ID")
	}

	currentUserID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)
	csrfToken := local.GetCSRFToken(c)

	conv, err := message.GetConversation(conversationID, currentUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	if conv.OwnerID != currentUserID && conv.InquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Only conversation participants can remove rocks")
	}

	err = rock.UnthrowRock(currentUserID, conversationID)
	if err != nil {
		if err == rock.ErrRockNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Rock not found")
		}
		logger.Error("Failed to unthrow rock", "error", err, "conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove rock")
	}
	if err := rockopinion.Invalidate(conversationID); err != nil {
		logger.Error("Failed to invalidate rock opinion",
			"error", err, "conversationID", conversationID)
	}

	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get conversation")
	}

	return renderConversationModalSwapOOB(c, conv, currentUserID, tz, csrfToken)
}

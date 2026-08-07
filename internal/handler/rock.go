package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/param"
	"github.com/rocky-ads/site/internal/rock"
	"github.com/rocky-ads/site/internal/rockopinion"
	"github.com/rocky-ads/site/internal/user"
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
		return fiber.NewError(fiber.StatusForbidden,
			"Only conversation participants can throw rocks")
	}
	if !message.MessagingAllowed(conv) {
		return fiber.NewError(fiber.StatusBadRequest,
			"Messaging is closed for this conversation")
	}

	if err := rock.ThrowRock(currentUserID, conversationID); err != nil {
		return throwRockError(err, conversationID, currentUserID)
	}

	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			"Failed to get conversation")
	}

	sendRockEventSSE(conv, currentUserID)
	return renderConversationRockEventAppend(c, conv, currentUserID, tz,
		csrfToken)
}

// ThrowRockOnAdHandler creates the conversation if needed, then throws a rock.
// Used when the inquirer opens a new conversation modal (conversation ID 0).
func ThrowRockOnAdHandler(c *fiber.Ctx) error {
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
	if !a.IsActive() {
		return fiber.NewError(fiber.StatusBadRequest,
			"Cannot throw a rock at an inactive or deleted ad")
	}
	if !user.Exists(a.UserID) {
		return fiber.NewError(fiber.StatusBadRequest,
			"This account is no longer available")
	}
	if a.UserID == currentUserID {
		return fiber.NewError(fiber.StatusForbidden,
			"You cannot throw a rock at your own ad")
	}

	conv, err := message.GetConversationByAdAndInquirer(
		adID, a.UserID, currentUserID)
	if err != nil {
		if err != message.ErrConversationNotFound {
			return fiber.NewError(fiber.StatusInternalServerError,
				"Failed to get conversation")
		}
		conv, err = message.CreateConversation(adID, a.UserID, currentUserID)
		if err != nil {
			logger.Error("Failed to create conversation for rock throw",
				"error", err, "adID", adID, "ownerID", a.UserID,
				"inquirerID", currentUserID)
			return fiber.NewError(fiber.StatusInternalServerError,
				"Failed to create conversation")
		}
	}

	if !message.MessagingAllowed(conv) {
		return fiber.NewError(fiber.StatusBadRequest,
			"Messaging is closed for this conversation")
	}

	if err := rock.ThrowRock(currentUserID, conv.ID); err != nil {
		return throwRockError(err, conv.ID, currentUserID)
	}

	conv, err = message.GetConversationByID(conv.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			"Failed to get conversation")
	}

	sendRockEventSSE(conv, currentUserID)

	view, err := message.BuildConversationModal(conv, currentUserID, tz)
	if err != nil {
		return buildConversationModalError(err)
	}
	return renderConversationModalView(c, view, currentUserID, tz, csrfToken,
		"conversation-0-modal", modalRenderSwapOOB)
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
		return fiber.NewError(fiber.StatusForbidden,
			"Only conversation participants can remove rocks")
	}

	err = rock.UnthrowRock(currentUserID, conversationID)
	if err != nil {
		if err == rock.ErrRockNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Rock not found")
		}
		logger.Error("Failed to unthrow rock", "error", err,
			"conversationID", conversationID, "userID", currentUserID)
		return fiber.NewError(fiber.StatusInternalServerError,
			"Failed to remove rock")
	}
	if err := rockopinion.Invalidate(conversationID); err != nil {
		logger.Error("Failed to invalidate rock opinion",
			"error", err, "conversationID", conversationID)
	}

	conv, err = message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			"Failed to get conversation")
	}

	sendRockEventSSE(conv, currentUserID)
	return renderConversationRockEventAppend(c, conv, currentUserID, tz,
		csrfToken)
}

func throwRockError(err error, conversationID, userID int) error {
	if err == rock.ErrMaxRocksReached {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf(
			"You have reached the maximum of %d outstanding rocks",
			config.MaxOutstandingRocks,
		))
	}
	if err == rock.ErrRockAlreadyThrown {
		return fiber.NewError(fiber.StatusBadRequest,
			"An rock has already been thrown at this conversation")
	}
	logger.Error("Failed to throw rock", "error", err,
		"conversationID", conversationID, "userID", userID)
	return fiber.NewError(fiber.StatusInternalServerError,
		"Failed to throw rock")
}

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/rock"
	"github.com/rocky-ads/site/ui"
)

func UserMenuHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	userName := local.GetUserName(c)
	isAdmin := local.GetUserIsAdmin(c)
	hasUnread, _ := message.GetHasUnread(userID)
	rockCount, _ := rock.GetUserRockCount(userID)
	userRockCount, _ := rock.GetRockCountForUser(userID)
	return render(c, ui.UserMenu(userName, userID, isAdmin, hasUnread, rockCount, userRockCount))
}

func UserMyAdsHandler(c *fiber.Ctx) error {
	return userMyAdsTabHandler(c, "active")
}

func UserMyAdsTabHandler(c *fiber.Ctx) error {
	tabID := c.Params("tab")
	if tabID != "bookmarked" && tabID != "active" && tabID != "deleted" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid tab")
	}
	return userMyAdsTabHandler(c, tabID)
}

func userMyAdsTabHandler(c *fiber.Ctx, activeTab string) error {
	userID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	// Get ad IDs for the selected tab
	adIDs, err := ad.GetUserAdIDs(userID, activeTab)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load ads")
	}

	// Render ads using list view (no pagination for My Ads)
	adNodes, err := ad.AdNodes(adIDs, userID, ui.ViewList, 1, loc, csrfToken, false)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to render ads")
	}

	// Check if this is a tab switch (HTMX request) or full page load
	if c.Get("HX-Request") != "" {
		return render(c, ui.MyAdsContainer(activeTab, adNodes))
	}

	return renderPage(c, "My Ads", ui.MyAdsPage(activeTab, adNodes))
}

func UserSettingsHandler(c *fiber.Ctx) error {
	return renderPage(c, "Settings", ui.SettingsPage())
}

func UserAboutHandler(c *fiber.Ctx) error {
	return renderPage(c, "About", ui.AboutPage())
}

func UserRockConversationHandler(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	ordinal, err := c.ParamsInt("ordinal")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid rock ordinal")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	// Get conversation ID by ordinal
	conversationID, err := rock.GetConversationIDForUserRockByOrdinal(userID, ordinal)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Rock conversation not found")
	}

	// Get conversation
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
		}
	}

	return renderConversationModal(c, conv, currentUserID, loc, csrfToken)
}

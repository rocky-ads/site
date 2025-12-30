package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/ui"
)

func UserMenuHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	userName := local.GetUserName(c)
	isAdmin := local.GetUserIsAdmin(c)
	messageCount := GetMessageCount(userID)
	return render(c, ui.UserMenu(userName, isAdmin, messageCount))
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

func UserMessagesHandler(c *fiber.Ctx) error {
	return renderPage(c, "Messages", ui.MessagesPage())
}

func UserSettingsHandler(c *fiber.Ctx) error {
	return renderPage(c, "Settings", ui.SettingsPage())
}

func UserAboutHandler(c *fiber.Ctx) error {
	return renderPage(c, "About", ui.AboutPage())
}

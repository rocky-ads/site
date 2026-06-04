package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/egg"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/ui"
	"github.com/rocky-ads/site/user"
)

const (
	memberSinceLayoutPage    = "January 2, 2006"
	memberSinceLayoutSummary = "Jan 2006"
)

func userProfileData(u user.User, activeAdCount, userEggCount int, loc *time.Location, memberSinceLayout string) ui.UserProfileData {
	return ui.UserProfileData{
		Name:          u.Name,
		MemberSince:   u.CreatedAt.In(loc).Format(memberSinceLayout),
		ActiveAdCount: activeAdCount,
		UserEggCount:  userEggCount,
	}
}

func loadUserMenuContext(userID int) (name string, isAdmin, hasUnread bool, eggCount, userEggCount int, err error) {
	u, err := user.GetByID(userID)
	if err != nil {
		return "", false, false, 0, 0, err
	}
	hasUnread, _ = message.GetHasUnread(userID)
	eggCount, _ = egg.GetUserEggCount(userID)
	userEggCount, _ = egg.GetEggCountForUser(userID)
	return u.Name, u.IsAdmin, hasUnread, eggCount, userEggCount, nil
}

func UserMenuHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	name, isAdmin, hasUnread, eggCount, userEggCount, err := loadUserMenuContext(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load user menu")
	}
	return render(c, ui.UserMenu(name, userID, isAdmin, hasUnread, eggCount, userEggCount))
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

	ads, err := ad.GetAds(userID, adIDs, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load ads")
	}

	adNodes := ui.AdNodes(adCardsFrom(ads, loc), userID, ui.ViewList, 1, csrfToken, false)

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

func UserProfileHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}
	u, err := user.GetByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}
	activeAdCount, _ := ad.CountActiveAdsByUser(id)
	userEggCount, _ := egg.GetEggCountForUser(id)
	loc := cookie.GetLocation(c)
	d := userProfileData(u, activeAdCount, userEggCount, loc, memberSinceLayoutPage)
	return renderPage(c, u.Name, ui.UserProfilePage(d))
}

func UserSummaryHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}
	u, err := user.GetByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}
	activeAdCount, _ := ad.CountActiveAdsByUser(id)
	userEggCount, _ := egg.GetEggCountForUser(id)
	loc := cookie.GetLocation(c)
	d := userProfileData(u, activeAdCount, userEggCount, loc, memberSinceLayoutSummary)
	return render(c, ui.UserSummaryFragment(d))
}

func UserEggConversationHandler(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	ordinal, err := c.ParamsInt("ordinal")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid egg ordinal")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	// Get conversation ID by ordinal
	conversationID, err := egg.GetConversationIDForUserEggByOrdinal(userID, ordinal)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Egg conversation not found")
	}

	conv, _, err := message.OpenConversation(conversationID, currentUserID)
	if errors.Is(err, message.ErrModalAccess) {
		return fiber.NewError(fiber.StatusForbidden, "Conversation not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	return renderConversationModal(c, conv, currentUserID, loc, csrfToken)
}

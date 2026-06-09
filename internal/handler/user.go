package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/egg"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/password"
	"github.com/rocky-ads/site/internal/ui"
	"github.com/rocky-ads/site/internal/user"
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

func loadUserMenuContext(userID int, loc *time.Location) (name, memberSince string, isAdmin, hasUnread bool, eggCount, userEggCount int, err error) {
	u, err := user.GetByID(userID)
	if err != nil {
		return "", "", false, false, 0, 0, err
	}
	hasUnread, _ = message.GetHasUnread(userID)
	eggCount, _ = egg.GetUserEggCount(userID)
	userEggCount, _ = egg.GetEggCountForUser(userID)
	memberSince = u.CreatedAt.In(loc).Format(memberSinceLayoutSummary)
	return u.Name, memberSince, u.IsAdmin, hasUnread, eggCount, userEggCount, nil
}

func UserMenuHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	name, memberSince, isAdmin, hasUnread, eggCount, userEggCount, err := loadUserMenuContext(userID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to load user menu")
	}
	return render(c, ui.UserMenu(name, memberSince, userID, isAdmin, hasUnread, eggCount, userEggCount))
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

	adNodes := ui.AdNodes(adCardsFrom(ads, userID, loc), userID, ui.ViewList, 1, csrfToken, false)

	// Check if this is a tab switch (HTMX request) or full page load
	if c.Get("HX-Request") != "" {
		return render(c, ui.MyAdsContainer(activeTab, adNodes))
	}

	return renderPage(c, "My Ads", ui.MyAdsPage(activeTab, adNodes))
}

func UserSettingsHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	u, err := user.GetByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}
	return renderPage(c, "Settings", ui.SettingsPage(u.Name, u.PhoneE64, u.SMSOptedOut))
}

func NotificationsToggleHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	enabled := c.FormValue("enabled") == "true"
	if err := user.SetSMSOptOut(userID, !enabled); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update notification preferences")
	}
	return render(c, ui.NotificationsSection(!enabled))
}

func ChangePasswordHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	if err := password.ValidatePasswordChange(currentPassword, newPassword, confirmPassword); err != nil {
		return showErrorTo(c, ui.SettingsChangePasswordErrorID, err.Error())
	}

	u, err := user.GetByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if !password.VerifyPassword(currentPassword, u.PasswordHash, u.PasswordSalt) {
		return showErrorTo(c, ui.SettingsChangePasswordErrorID, "Invalid current password")
	}

	newHash, newSalt, err := password.HashPassword(newPassword)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to change password")
	}

	if err := user.UpdatePassword(userID, newHash, newSalt, "argon2id"); err != nil {
		logger.Error("Failed to update password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to change password")
	}

	logout(c)
	c.Set("HX-Redirect", "/login")
	return c.SendStatus(fiber.StatusOK)
}

func DeleteAccountHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	passwd := c.FormValue("password")
	if passwd == "" {
		return showErrorTo(c, ui.SettingsDeleteAccountErrorID, "Password is required")
	}

	u, err := user.GetByID(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	if !password.VerifyPassword(passwd, u.PasswordHash, u.PasswordSalt) {
		return showErrorTo(c, ui.SettingsDeleteAccountErrorID, "Invalid password")
	}

	if err := user.DeleteUser(userID); err != nil {
		logger.Error("Failed to delete account", "error", err, "userID", userID)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete account")
	}

	logout(c)
	c.Set("HX-Redirect", "/")
	return c.SendStatus(fiber.StatusOK)
}

func AboutHandler(c *fiber.Ctx) error {
	return renderPage(c, "About", ui.AboutPage())
}

func FAQHandler(c *fiber.Ctx) error {
	return renderPage(c, "FAQ", ui.FAQPage())
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

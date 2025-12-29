package handler

import (
	"github.com/gofiber/fiber/v2"
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

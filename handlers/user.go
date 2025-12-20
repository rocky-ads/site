package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/ui"
)

func UserMenuHandler(c *fiber.Ctx) error {
	userName := local.GetUserName(c)
	isAdmin := local.GetUserIsAdmin(c)
	return render(c, ui.UserMenu(userName, isAdmin))
}

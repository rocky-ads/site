package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/local"
)

// redirectToLogin redirects to the login page
func redirectToLogin(c *fiber.Ctx) error {
	return c.Redirect("/login", fiber.StatusFound)
}

// AuthRequired is a middleware that requires a user to be logged in.
func AuthRequired(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	if userID == 0 {
		return redirectToLogin(c)
	}
	return c.Next()
}

// AdminRequired is a middleware that requires a user to be an admin.
func AdminRequired(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}
	if !local.GetUserIsAdmin(c) {
		return c.Status(fiber.StatusForbidden).SendString("Forbidden")
	}
	return c.Next()
}

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/local"
)

// redirectToLogin redirects to the login page
// For HTMX requests, uses HX-Redirect header to trigger a full page redirect
func redirectToLogin(c *fiber.Ctx) error {
	// Check if this is an HTMX request
	if c.Get("HX-Request") != "" {
		// Use HX-Redirect header for HTMX requests to trigger full page redirect
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	// Regular redirect for non-HTMX requests
	return c.Redirect("/login", fiber.StatusFound)
}

// AuthRequired is a middleware that requires a user to be logged in.
func AuthRequired(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	if !local.IsLoggedIn(userID) {
		return redirectToLogin(c)
	}
	return c.Next()
}

// AdminRequired is a middleware that requires a user to be an admin.
func AdminRequired(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	if !local.IsLoggedIn(userID) {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}
	if !local.GetUserIsAdmin(c) {
		return c.Status(fiber.StatusForbidden).SendString("Forbidden")
	}
	return c.Next()
}

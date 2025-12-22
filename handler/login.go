package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/password"
	"github.com/rocky-ads/site/ui"
	"github.com/rocky-ads/site/user"
)

func LoginHandler(c *fiber.Ctx) error {
	return renderPage(c, "Login", ui.LoginPage())
}

func logout(c *fiber.Ctx) {
	cookie.ClearJWT(c)
	local.SetUserID(c, 0)
	local.SetUserName(c, "")
	local.SetUserIsAdmin(c, false)
}

func LogoutHandler(c *fiber.Ctx) error {
	logout(c)
	return renderPage(c, "Logout", ui.LogoutPage())
}

func LoginSubmitHandler(c *fiber.Ctx) error {
	userName := c.FormValue("username")
	passwd := c.FormValue("password")

	logger.Info("Login attempt", "userName", userName)

	// Validate input
	if userName == "" {
		return showError(c, "Username is required")
	}
	if passwd == "" {
		return showError(c, "Password is required")
	}

	// Get user from database
	u, err := user.GetByName(userName)
	if err != nil {
		// User not found or error - don't reveal which
		return showError(c, "Invalid username or password")
	}

	// Verify password
	if !password.VerifyPassword(passwd, u.PasswordHash, u.PasswordSalt) {
		return showError(c, "Invalid username or password")
	}

	// Generate JWT token
	token, err := generateJWTToken(&u)
	if err != nil {
		logger.Error("Failed to generate token", "error", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Set JWT cookie
	cookie.SetJWT(c, token)

	// Redirect to home page using HTMX
	c.Set("HX-Redirect", "/")
	return c.SendStatus(fiber.StatusOK)
}

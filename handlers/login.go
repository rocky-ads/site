package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/password"
	"github.com/rocky-ads/site/ui"
	"github.com/rocky-ads/site/user"
)

func LoginHandler(c *fiber.Ctx) error {
	return renderPage(c, "Login", ui.LoginPage())
}

func LoginSubmitHandler(c *fiber.Ctx) error {
	userName := c.FormValue("username")
	passwd := c.FormValue("password")

	logger.Info("Login attempt", "userName", userName)

	// Validate input
	if userName == "" {
		return showLoginError(c, "Username is required")
	}
	if passwd == "" {
		return showLoginError(c, "Password is required")
	}

	// Get user from database
	u, err := user.GetByName(userName)
	if err != nil {
		// User not found or error - don't reveal which
		return showLoginError(c, "Invalid username or password")
	}

	// Verify password
	if !password.VerifyPassword(passwd, u.PasswordHash, u.PasswordSalt) {
		return showLoginError(c, "Invalid username or password")
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

func showLoginError(c *fiber.Ctx, errMsg string) error {
	return render(c, ui.ErrorDiv(errMsg))
}

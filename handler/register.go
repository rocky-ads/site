package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ui"
)

func RegisterHandler(c *fiber.Ctx) error {
	logout(c)
	return renderPage(c, "Register", ui.RegisterPage())
}

func RegisterSubmitHandler(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

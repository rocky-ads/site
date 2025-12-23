package handler

import (
	"github.com/gofiber/fiber/v2"
)

func ViewHandler(c *fiber.Ctx) error {
	view := c.Params("view")
	view = view
	return nil
}

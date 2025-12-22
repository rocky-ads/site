package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ui"
)

func showError(c *fiber.Ctx, errMsg string) error {
	return render(c, ui.ErrorDiv(errMsg))
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	c.Status(code)
	title := fmt.Sprintf("%d - %s", code, message)
	return renderPage(c, title, ui.ErrorPageContent(code, message))
}

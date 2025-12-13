package handlers

import (
	"github.com/rocky-ads/site/ui"

	"github.com/gofiber/fiber/v2"
)

// HomeHandler handles the homepage route
func HomeHandler(c *fiber.Ctx) error {
	return renderPage(c, "Rocky Ads - Classified ads without the newspaper",
		ui.HomePageContent())
}

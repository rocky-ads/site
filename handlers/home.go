package handlers

import (
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/ui"

	"github.com/gofiber/fiber/v2"
)

func HomeHandler(c *fiber.Ctx) error {
	categoryID, err := cookie.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categoryName, err := ad.GetCategoryName(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categoryImage, err := ad.GetCategoryImageFile(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return renderPage(c, "Rocky Ads - Classified ads without the newspaper",
		ui.HomePage(categoryName, categoryImage))
}

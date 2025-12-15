package handlers

import (
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/services"
	"github.com/rocky-ads/site/ui"

	"github.com/gofiber/fiber/v2"
)

func HomeHandler(c *fiber.Ctx) error {
	categoryID, fiberErr := cookie.GetCategoryID(c)
	if fiberErr != nil {
		return fiberErr
	}

	categoryName, err := services.GetCategoryNameByID(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categoryImage, err := services.GetCategoryImageFile(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return renderPage(c, "Rocky Ads - Classified ads without the newspaper",
		ui.HomePage(categoryName, categoryImage))
}

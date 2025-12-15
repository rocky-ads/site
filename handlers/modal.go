package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/services"
	"github.com/rocky-ads/site/ui"
)

func CategorySelectHandler(c *fiber.Ctx) error {
	categoryID, err := cookie.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categories, err := services.GetCategories()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return render(c, ui.CategorySelectModal(categoryID, categories))
}

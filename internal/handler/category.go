package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/ui"
	g "maragu.dev/gomponents"
)

func CategorySelectHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)
	returnParam := c.Query("return")

	categories, err := ad.GetCategories()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	options := make([]ui.CategoryOption, len(categories))
	for i, cat := range categories {
		options[i] = categoryOption(cat)
	}

	return render(c, ui.CategorySelectModal(categoryID, returnParam, options))
}

func ModalRemoveHandler(c *fiber.Ctx) error {
	name := c.Params("name")
	return render(c, g.Group(ui.RemoveModal(name)))
}

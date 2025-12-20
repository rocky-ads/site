package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

func CategorySelectHandler(c *fiber.Ctx) error {
	categoryID, err := cookie.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categories, err := ad.GetCategories()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var categoryItems []g.Node
	for _, cat := range categories {
		categoryItems = append(categoryItems, ui.CategoryItem(categoryID, cat.ID, cat.Name, cat.ImageFile))
	}

	return render(c, ui.CategorySelectModal(categoryItems))
}

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/ui"
)

func ViewHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	viewStr := c.Params("view")
	view := ui.ValidateView(viewStr)

	categoryID, err := cookie.GetCategoryID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	results, err := searchAndRenderAds(categoryID, make(field.Values))
	if err != nil {
		return err
	}

	cookie.SetView(c, view)

	return render(c, ui.SearchView(userID, view, results))
}

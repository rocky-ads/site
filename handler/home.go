package handler

import (
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/ui"

	"github.com/gofiber/fiber/v2"
)

func HomeHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)

	categoryName, err := ad.GetCategoryName(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categoryImage, err := ad.GetCategoryImageFile(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	userID := local.GetUserID(c)
	view := cookie.GetView(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	p := parseSearchParams(c, categoryID)
	results, err := searchAndRenderAds(p, userID, view, loc, csrfToken)
	if err != nil {
		return err
	}

	state := cookie.GetSearchState(c)
	return renderPage(c, config.ServerName,
		ui.HomePage(userID, view, categoryName, categoryImage,
			state.Q, state.Expanded, parseSearchFilters(c), results))
}

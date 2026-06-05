package handler

import (
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/ui"

	"github.com/gofiber/fiber/v2"
)

func HomeHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)

	category, err := ad.GetCategory(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	userID := local.GetUserID(c)
	view := ui.ValidateView(cookie.GetView(c))
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	p := parseSearchParams(c, categoryID)
	results, err := searchAndRenderAds(p, userID, view, loc, csrfToken)
	if err != nil {
		return err
	}

	state := cookie.GetSearchState(c)
	return renderPage(c, config.ServerName,
		ui.HomePage(userID, view, state.Q, state.Expanded,
			category, parseSearchFilters(c), results))
}

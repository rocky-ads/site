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
	tz := cookie.GetTimezone(c)
	csrfToken := local.GetCSRFToken(c)
	searchState := cookie.GetSearchState(c)
	unit := cookie.GetDistanceUnit(c)

	p := parseSearchParamsFromState(c, searchState, categoryID)
	results, err := searchAndRenderAds(
		p, userID, view, tz, csrfToken,
		searchLocationDisplay(searchState.Location), searchState.Within, unit,
	)
	if err != nil {
		return err
	}

	return renderPage(c, config.ServerName,
		ui.HomePage(userID, view, searchState.Q, searchState.Expanded, searchVisible(searchState),
			categoryOption(category), filterableFacets(category),
			searchStateToFilters(searchState, unit), results))
}

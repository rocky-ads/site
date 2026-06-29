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

	category := ad.GetCategory(cookie.GetCategoryID(c))
	userID := local.GetUserID(c)
	view := ui.ValidateView(cookie.GetView(c))
	tz := cookie.GetTimezone(c)
	csrfToken := local.GetCSRFToken(c)
	distanceUnit := cookie.GetDistanceUnit(c)
	searchState := cookie.GetSearchState(c)
	page := c.QueryInt("page", 1)

	p := parseSearchParams(searchState, page, category.ID, userID, distanceUnit, tz)

	results, err := searchAndRenderAds(
		p, userID, view, tz, csrfToken,
		searchLocationDisplay(searchState.Location), searchState.Within, distanceUnit,
	)
	if err != nil {
		return err
	}

	return renderPage(c, config.ServerName,
		ui.HomePage(userID, view, searchState.Q, searchState.Expanded, searchVisible(searchState),
			categoryOption(category), filterableFacets(category),
			searchStateToFilters(searchState, distanceUnit), results))
}

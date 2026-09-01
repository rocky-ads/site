package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/ui"
)

func HomeHandler(c *fiber.Ctx) error {

	category := ad.GetCategory(cookie.GetCategoryID(c))
	userID := local.GetUserID(c)
	view := ui.ValidateView(cookie.GetView(c))
	searchState := cookie.GetSearchState(c)
	distanceUnit := cookie.GetDistanceUnit(c)

	_, results, err := searchAndRender(c, searchState, category.ID, view)
	if err != nil {
		return err
	}

	return renderPage(c, "Classified Ads",
		ui.HomePage(userID, view, searchState.Q, searchState.Expanded,
			searchVisible(searchState), categoryOption(category),
			filterFacetsForExpanded(category, searchState.Expanded),
			searchStateToFilters(searchState, distanceUnit),
			results))
}

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/ui"
)

func ViewHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	viewStr := c.Params("view")
	view := ui.ValidateView(viewStr)
	tz := cookie.GetTimezone(c)
	categoryID := cookie.GetCategoryID(c)
	csrfToken := local.GetCSRFToken(c)
	distanceUnit := cookie.GetDistanceUnit(c)
	searchState := cookie.GetSearchState(c)
	page := c.QueryInt("page", 1)
	p := parseSearchParams(searchState, page, categoryID, userID, distanceUnit, tz)
	results, err := searchAndRenderAds(
		p, userID, view, tz, csrfToken,
		searchLocationDisplay(searchState.Location), searchState.Within, distanceUnit,
	)
	if err != nil {
		return err
	}

	cookie.SetView(c, view)

	filters := searchStateToFilters(searchState, distanceUnit)
	return render(c, ui.SearchView(view, filters, results))
}

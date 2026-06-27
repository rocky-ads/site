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

	p := parseSearchParams(c, categoryID)
	state := cookie.GetSearchState(c)
	results, err := searchAndRenderAds(
		p, userID, view, tz, csrfToken,
		searchLocationDisplay(state.Location), state.Within, cookie.GetDistanceUnit(c),
	)
	if err != nil {
		return err
	}

	cookie.SetView(c, view)

	filters := searchStateToFilters(state, cookie.GetDistanceUnit(c))
	return render(c, ui.SearchView(view, filters, results))
}

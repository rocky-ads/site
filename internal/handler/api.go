package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/param"
	"github.com/rocky-ads/site/internal/ui"
	g "maragu.dev/gomponents"
)

func SwitchCategoryHandler(c *fiber.Ctx) error {
	redirect := categorySwitchRedirect(c)

	categoryID := param.GetCategoryID(c)
	cookie.SetCategoryID(c, categoryID)

	// Search state is owned by the home page search widget; other pages
	// (new ad, edit ad, etc.) only change the category cookie.
	if redirect == "/" {
		switchCategoryHome(c, categoryID)
	}

	if c.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", redirect)
		return render(c, g.Group(ui.RemoveModal("category")))
	}
	return c.Redirect(redirect, fiber.StatusFound)
}

func ShortCategoryHandler(c *fiber.Ctx) error {
	categoryID := param.GetCategoryID(c)
	cookie.SetCategoryID(c, categoryID)
	switchCategoryHome(c, categoryID)
	return c.Redirect("/", fiber.StatusFound)
}

func switchCategoryHome(c *fiber.Ctx, categoryID int) {
	state := saveSearchStateFromRequest(c, nil, true)
	state = clearFacetFilters(state, categoryID)
	cookie.SetSearchState(c, state)
}

func categorySwitchRedirect(c *fiber.Ctx) string {
	if redirect := safeReturnPath(c.Query("return")); redirect != "" {
		return redirect
	}
	return "/"
}

func renderFilterPanelResponse(c *fiber.Ctx, state cookie.SearchState) error {
	view, _, results, err := searchFromRequest(c, state)
	if err != nil {
		return err
	}

	distanceUnit := cookie.GetDistanceUnit(c)
	category := ad.GetCategory(cookie.GetCategoryID(c))
	visible := searchVisible(state)
	filterFacets := filterFacetsForExpanded(category, state.Expanded)
	filters := searchStateToFilters(state, distanceUnit)

	return render(c, g.Group([]g.Node{
		ui.SearchAreaOOB(state.Q, state.Expanded, filterFacets, filters, visible),
		ui.SearchToggleOOB(visible),
		ui.SearchResultsOOB(view, results),
	}))
}

func ShowFiltersHandler(c *fiber.Ctx) error {
	expanded := true
	state := saveSearchStateFromRequest(c, &expanded, false)
	state.SearchOpen = true
	cookie.SetSearchState(c, state)
	return renderFilterPanelResponse(c, state)
}

func HideFiltersHandler(c *fiber.Ctx) error {
	expanded := false
	state := saveSearchStateFromRequest(c, &expanded, true)
	return renderFilterPanelResponse(c, state)
}

func ToggleSearchHandler(c *fiber.Ctx) error {
	state := cookie.GetSearchState(c)
	state.SearchOpen = true
	cookie.SetSearchState(c, state)
	return renderFilterPanelResponse(c, state)
}

func SearchPageHandler(c *fiber.Ctx) error {
	state := saveSearchStateFromRequest(c, nil, true)
	view, page, results, err := searchFromRequest(c, state)
	if err != nil {
		return err
	}
	return render(c, ui.SearchResultsResponse(view, page, results))
}

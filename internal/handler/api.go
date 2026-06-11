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
	state := saveSearchStateFromRequest(c, nil, true)

	categoryID := param.GetCategoryID(c)

	if _, err := ad.GetCategory(categoryID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	cookie.SetCategoryID(c, categoryID)

	state = clearFacetFilters(state, categoryID)
	cookie.SetSearchState(c, state)

	redirect := c.Query("return")
	if redirect == "" || redirect[0] != '/' || (len(redirect) > 1 && redirect[1] == '/') {
		redirect = "/"
	}
	c.Set("HX-Redirect", redirect)
	return c.Send(nil)
}

func renderFilterPanelResponse(c *fiber.Ctx, state cookie.SearchState, panel g.Node) error {
	view, _, results, err := searchFromRequest(c, state)
	if err != nil {
		return err
	}

	return render(c, g.Group([]g.Node{
		panel,
		ui.SearchBarOOB(state.Q, state.Expanded),
		ui.SearchResultsOOB(view, results),
	}))
}

func ShowFiltersHandler(c *fiber.Ctx) error {
	expanded := true
	state := saveSearchStateFromRequest(c, &expanded, false)
	unit := distanceUnit(c)
	categoryID := cookie.GetCategoryID(c)
	category, err := ad.GetCategory(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return renderFilterPanelResponse(c, state, ui.FilterPanel(state.Q, filterableFacets(category), searchStateToFilters(state, unit)))
}

func HideFiltersHandler(c *fiber.Ctx) error {
	expanded := false
	state := saveSearchStateFromRequest(c, &expanded, true)
	return renderFilterPanelResponse(c, state, g.Text(""))
}

func SearchPageHandler(c *fiber.Ctx) error {
	state := saveSearchStateFromRequest(c, nil, true)
	view, page, results, err := searchFromRequest(c, state)
	if err != nil {
		return err
	}
	return render(c, ui.SearchResultsResponse(view, page, results))
}

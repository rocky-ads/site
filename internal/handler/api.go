package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/facet"
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

func renderFilterPanelResponse(c *fiber.Ctx, state cookie.SearchState) error {
	view, _, results, err := searchFromRequest(c, state)
	if err != nil {
		return err
	}

	unit := distanceUnit(c)
	categoryID := cookie.GetCategoryID(c)
	category, err := ad.GetCategory(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	var filterFacets []facet.Def
	if state.Expanded {
		filterFacets = filterableFacets(category)
	}

	return render(c, g.Group([]g.Node{
		ui.SearchAreaOOB(state.Q, state.Expanded, filterFacets, searchStateToFilters(state, unit)),
		ui.SearchResultsOOB(view, results),
	}))
}

func ShowFiltersHandler(c *fiber.Ctx) error {
	expanded := true
	state := saveSearchStateFromRequest(c, &expanded, false)
	return renderFilterPanelResponse(c, state)
}

func HideFiltersHandler(c *fiber.Ctx) error {
	expanded := false
	state := saveSearchStateFromRequest(c, &expanded, true)
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

package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/ui"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
	g "maragu.dev/gomponents"
)

func SearchLocationModalHandler(c *fiber.Ctx) error {
	unit := cookie.GetDistanceUnit(c)
	state := cookie.GetSearchState(c)
	filters := searchStateToFilters(state, unit)
	return render(c, ui.SearchLocationModal(filters))
}

func SearchLocationSaveHandler(c *fiber.Ctx) error {
	state := saveSearchStateFromRequest(c, nil, true)

	view, _, results, err := searchFromRequest(c, state)
	if err != nil {
		return err
	}

	unit := cookie.GetDistanceUnit(c)
	filters := searchStateToFilters(state, unit)

	nodes := []g.Node{
		uiads.SearchLocationOOB(filters),
		ui.SearchResultsOOB(view, results),
	}
	nodes = append(nodes, ui.RemoveModal("search-location")...)
	return render(c, g.Group(nodes))
}

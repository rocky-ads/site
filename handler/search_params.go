package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/location"
	"github.com/rocky-ads/site/search"
	uiads "github.com/rocky-ads/site/ui/ads"
	"github.com/rocky-ads/site/user"
)

func distanceUnit(c *fiber.Ctx) string {
	if unit, ok := local.GetDistanceUnit(c); ok {
		return unit
	}

	userID := local.GetUserID(c)
	unit := location.DistanceUnitFromTimezone(c.Cookies("timezone"))
	if userID != 0 {
		u, err := user.GetByID(userID)
		if err == nil && u.PhoneE64 != "" {
			unit = location.DistanceUnitFromPhone(u.PhoneE64)
		}
	}

	local.SetDistanceUnit(c, unit)
	return unit
}

func searchStateToFilters(state cookie.SearchState, unit string) uiads.SearchFilters {
	return uiads.SearchFilters{
		PriceMin:   state.PriceMin,
		PriceMax:   state.PriceMax,
		Location:   state.Location,
		Radius:     state.Radius,
		RadiusUnit: unit,
	}
}

func parseSearchFilters(c *fiber.Ctx) uiads.SearchFilters {
	return searchStateToFilters(cookie.GetSearchState(c), distanceUnit(c))
}

func parseSearchParams(c *fiber.Ctx, categoryID int) search.Params {
	return parseSearchParamsFromState(c, cookie.GetSearchState(c), categoryID)
}

func pageLimitOffset(c *fiber.Ctx) (limit, offset int) {
	page := c.QueryInt("page", 1)
	limit = config.SearchPageSize
	offset = (page - 1) * limit
	return limit, offset
}

func parseSearchParamsFromState(c *fiber.Ctx, state cookie.SearchState, categoryID int) search.Params {
	limit, offset := pageLimitOffset(c)
	p := search.Params{
		CategoryID: categoryID,
		Limit:      limit,
		Offset:     offset,
		Q:          state.Q,
	}
	if !state.Expanded {
		return p
	}

	unit := distanceUnit(c)
	f := searchStateToFilters(state, unit)
	p.PriceMin = f.PriceMin
	p.PriceMax = f.PriceMax

	if f.Location != "" && f.Radius > 0 {
		lat, lon, ok, err := location.ResolveLocation(f.Location)
		if err == nil && ok {
			p.CenterLat = lat
			p.CenterLon = lon
			if f.RadiusUnit == location.UnitKm {
				p.RadiusKm = float64(f.Radius)
			} else {
				p.RadiusKm = location.MilesToKm(float64(f.Radius))
			}
			p.HasGeo = true
		}
	}

	return p
}

// saveSearchStateFromRequest updates the search cookie and returns the new
// state (the request cookie is unchanged until the response is received).
func saveSearchStateFromRequest(c *fiber.Ctx, expanded *bool, fromForm bool) cookie.SearchState {
	unit := distanceUnit(c)
	state := cookie.GetSearchState(c)
	if fromForm {
		state.Q = strings.TrimSpace(c.Query("q"))
		if state.Expanded {
			state.PriceMin = parseOptionalAmount(c.Query("price_min"))
			state.PriceMax = parseOptionalAmount(c.Query("price_max"))
			state.Location = strings.TrimSpace(c.Query("location"))
			state.Radius = parseRadius(c.Query("radius"), unit)
			if state.Location == "" {
				state.Radius = 0
			}
		}
	}
	if expanded != nil {
		state.Expanded = *expanded
	}
	cookie.SetSearchState(c, state)
	return state
}

func parseOptionalAmount(raw string) *int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	amount, err := strconv.Atoi(raw)
	if err != nil || amount < 0 {
		return nil
	}
	return &amount
}

func parseRadius(raw, unit string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	opts := search.RadiusMileOptions
	if unit == location.UnitKm {
		opts = search.RadiusKmOptions
	}
	for _, opt := range opts {
		if n == opt {
			return n
		}
	}
	return 0
}

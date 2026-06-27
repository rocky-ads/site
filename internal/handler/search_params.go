package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/location"
	"github.com/rocky-ads/site/internal/search"
	"github.com/rocky-ads/site/internal/ui"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
)

func searchStateToFilters(state cookie.SearchState, unit string) uiads.SearchFilters {
	opts := search.WithinMileOptions
	if unit == location.UnitKm {
		opts = search.WithinKmOptions
	}
	return uiads.SearchFilters{
		Facets:          state.Facets,
		Location:        state.Location,
		LocationDisplay: searchLocationDisplay(state.Location),
		Within:          state.Within,
		WithinUnit:      unit,
		WithinOptions:   opts,
	}
}

func searchLocationDisplay(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	display, ok, err := location.DisplayTextForInput(raw)
	if err != nil || !ok || display == "" {
		return raw
	}
	return display
}

func categoryOption(cat ad.Category) ui.CategoryOption {
	return ui.CategoryOption{ID: cat.ID, Name: cat.Name, ImageFile: cat.ImageFile}
}

func filterableFacets(cat ad.Category) []facet.Def {
	defs := cat.Facets()
	out := make([]facet.Def, 0, len(defs))
	for _, d := range defs {
		if d.Filterable {
			out = append(out, d)
		}
	}
	return out
}

func parseSearchFilters(c *fiber.Ctx) uiads.SearchFilters {
	return searchStateToFilters(cookie.GetSearchState(c), cookie.GetDistanceUnit(c))
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
	unit := cookie.GetDistanceUnit(c)

	in := search.BuildInput{
		CategoryID: categoryID,
		Limit:      limit,
		Offset:     offset,
		Q:          state.Q,
		Location:   state.Location,
		Within:     state.Within,
		WithinUnit: unit,
	}
	if state.Expanded {
		in.FacetFilters = expandFacetFilters(state.Facets, cookie.GetTimezone(c))
	}

	p := search.BuildParams(in)
	p.UserID = local.GetUserID(c)
	p.Expanded = state.Expanded
	return p
}

// saveSearchStateFromRequest updates the search cookie and returns the new
// state (the request cookie is unchanged until the response is received).
func saveSearchStateFromRequest(c *fiber.Ctx, expanded *bool, fromForm bool) cookie.SearchState {
	unit := cookie.GetDistanceUnit(c)
	state := cookie.GetSearchState(c)
	if fromForm {
		state.Q = strings.TrimSpace(c.Query("q"))
		state.Location = strings.TrimSpace(c.Query("location"))
		state.Within = parseWithin(c.Query("within"), unit)
		if state.Location == "" {
			state.Within = 0
		}
		if state.Expanded {
			category, err := ad.GetCategory(cookie.GetCategoryID(c))
			if err == nil {
				state.Facets = parseFacetFilters(c, category)
			}
		}
	}
	if expanded != nil {
		state.Expanded = *expanded
	}
	cookie.SetSearchState(c, state)
	return state
}

func searchVisible(state cookie.SearchState) bool {
	if state.SearchOpen {
		return true
	}
	if strings.TrimSpace(state.Q) != "" || state.Expanded || len(state.Facets) > 0 {
		return true
	}
	return false
}

func parseFacetFilters(c *fiber.Ctx, category ad.Category) map[string]facet.Filter {
	filters := map[string]facet.Filter{}
	for _, d := range category.Facets() {
		if !d.Filterable {
			continue
		}
		switch d.Filter {
		case facet.FilterExact:
			val := strings.TrimSpace(c.Query(d.Key))
			if val != "" && d.ValidFilterValue(val) {
				filters[d.Key] = facet.Filter{Value: &val}
			}
		case facet.FilterCheckboxes:
			if vals := parseEnumCheckboxQuery(c, d.Key, d.Enum); len(vals) > 0 {
				filters[d.Key] = facet.Filter{Values: vals}
			}
		default:
			if d.Kind == facet.Date {
				min := parseOptionalDate(c.Query(d.Key + "_min"))
				max := parseOptionalDate(c.Query(d.Key + "_max"))
				if min != nil || max != nil {
					filters[d.Key] = facet.Filter{TextMin: min, TextMax: max}
				}
				continue
			}
			min := parseOptionalAmount(c.Query(d.Key + "_min"))
			max := parseOptionalAmount(c.Query(d.Key + "_max"))
			if min != nil || max != nil {
				filters[d.Key] = facet.Filter{Min: min, Max: max}
			}
		}
	}
	if len(filters) == 0 {
		return nil
	}
	return filters
}

func expandFacetFilters(
	filters map[string]facet.Filter,
	tz *time.Location,
) map[string]facet.Filter {
	if len(filters) == 0 {
		return filters
	}
	now := time.Now().In(tz)
	out := make(map[string]facet.Filter, len(filters))
	for k, f := range filters {
		out[k] = facet.ResolveFilterForSearch(k, f, now)
	}
	return out
}

func parseEnumCheckboxQuery(c *fiber.Ctx, key string, allowed []string) []string {
	allowedSet := make(map[string]bool, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = true
	}
	var vals []string
	c.Context().QueryArgs().VisitAll(func(k, v []byte) {
		if string(k) != key {
			return
		}
		s := strings.TrimSpace(string(v))
		if s == "" || !allowedSet[s] {
			return
		}
		for _, existing := range vals {
			if existing == s {
				return
			}
		}
		vals = append(vals, s)
	})
	return vals
}

func parseOptionalDate(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return nil
	}
	return &raw
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

func parseWithin(raw, unit string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	opts := search.WithinMileOptions
	if unit == location.UnitKm {
		opts = search.WithinKmOptions
	}
	for _, opt := range opts {
		if n == opt {
			return n
		}
	}
	return 0
}

func clearFacetFilters(state cookie.SearchState, categoryID int) cookie.SearchState {
	category, err := ad.GetCategory(categoryID)
	if err != nil {
		return state
	}
	allowed := make(map[string]bool, len(category.FacetKeys))
	for _, d := range category.Facets() {
		allowed[d.Key] = true
	}
	for key := range state.Facets {
		if !allowed[key] {
			delete(state.Facets, key)
		}
	}
	if len(state.Facets) == 0 {
		state.Facets = nil
	}
	return state
}

// SaveSearchStateForTest exposes saveSearchStateFromRequest for integration tests.
func SaveSearchStateForTest(c *fiber.Ctx, fromForm bool) cookie.SearchState {
	return saveSearchStateFromRequest(c, nil, fromForm)
}

// BuildSearchParamsForTest exposes parseSearchParamsFromState for integration tests.
func BuildSearchParamsForTest(c *fiber.Ctx, state cookie.SearchState, categoryID int) search.Params {
	return parseSearchParamsFromState(c, state, categoryID)
}

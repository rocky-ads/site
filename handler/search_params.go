package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/location"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/search"
	uiads "github.com/rocky-ads/site/ui/ads"
)

func parseSearchFilters(c *fiber.Ctx) uiads.SearchFilters {
	f := uiads.SearchFilters{}
	if v := strings.TrimSpace(c.Query("price_min")); v != "" {
		if amount, err := strconv.Atoi(v); err == nil && amount >= 0 {
			f.PriceMin = &amount
		}
	}
	if v := strings.TrimSpace(c.Query("price_max")); v != "" {
		if amount, err := strconv.Atoi(v); err == nil && amount >= 0 {
			f.PriceMax = &amount
		}
	}
	f.Location = strings.TrimSpace(c.Query("location"))
	f.RadiusMiles = parseRadiusMiles(c.Query("radius"))
	return f
}

func parseSearchParams(c *fiber.Ctx, categoryID int) search.Params {
	limit, offset := param.GetPageLimitOffset(c)
	f := parseSearchFilters(c)
	p := search.Params{
		CategoryID: categoryID,
		Limit:      limit,
		Offset:     offset,
		Q:          strings.TrimSpace(c.Query("q")),
		PriceMin:   f.PriceMin,
		PriceMax:   f.PriceMax,
	}

	if f.Location != "" && f.RadiusMiles > 0 {
		lat, lon, ok, err := location.ResolveLocation(f.Location)
		if err == nil && ok {
			p.CenterLat = lat
			p.CenterLon = lon
			p.RadiusKm = location.MilesToKm(float64(f.RadiusMiles))
			p.HasGeo = true
		}
	}

	return p
}

func parseRadiusMiles(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	for _, opt := range search.RadiusMileOptions {
		if n == opt {
			return n
		}
	}
	return 0
}

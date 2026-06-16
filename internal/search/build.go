package search

import (
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/location"
)

// BuildInput holds filter fields used to construct search Params.
type BuildInput struct {
	CategoryID   int
	Limit        int
	Offset       int
	Q            string
	Location     string
	Within       int
	WithinUnit   string
	FacetFilters map[string]facet.Filter
}

// BuildParams converts filter input into search Params, resolving geo when applicable.
func BuildParams(in BuildInput) Params {
	p := Params{
		CategoryID:   in.CategoryID,
		Limit:        in.Limit,
		Offset:       in.Offset,
		Q:            in.Q,
		FacetFilters: in.FacetFilters,
	}

	if in.Location == "" || in.Within <= 0 {
		return p
	}

	lat, lon, ok, err := location.ResolveLocation(in.Location)
	if err != nil || !ok {
		return p
	}

	p.CenterLat = lat
	p.CenterLon = lon
	if in.WithinUnit == location.UnitKm {
		p.WithinKm = float64(in.Within)
	} else {
		p.WithinKm = location.MilesToKm(float64(in.Within))
	}
	p.HasGeo = true
	return p
}

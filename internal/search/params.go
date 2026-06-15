package search

import "github.com/rocky-ads/site/internal/facet"

// RadiusMileOptions are valid search radius dropdown values (miles).
var RadiusMileOptions = []int{1, 5, 10, 25, 50, 100, 500}

// RadiusKmOptions are valid search radius dropdown values (kilometers).
var RadiusKmOptions = []int{1, 5, 10, 25, 50, 100, 500, 1000}

// Params holds hard filters for listing search.
type Params struct {
	CategoryID   int
	UserID       int
	Expanded     bool
	Limit        int
	Offset       int
	Q            string
	FacetFilters map[string]facet.Filter
	// CenterLat/CenterLon and RadiusKm apply when location text resolved and radius > 0.
	CenterLat float64
	CenterLon float64
	RadiusKm  float64
	HasGeo    bool
}

func (p Params) HasTextQuery() bool {
	return p.Q != ""
}

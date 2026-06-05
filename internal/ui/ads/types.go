package ads

import "github.com/rocky-ads/site/internal/facet"

// SearchFilters holds search filter panel state passed from handlers.
type SearchFilters struct {
	Facets        map[string]facet.Filter
	Location      string
	Radius        int
	RadiusUnit    string // "mi" or "km"
	RadiusOptions []int
}

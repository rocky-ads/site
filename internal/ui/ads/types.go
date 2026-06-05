package ads

import "github.com/rocky-ads/site/internal/facet"

// PriceRowView holds state for the new-ad price row (initial render and HTMX swaps).
type PriceRowView struct {
	IsFree   bool
	Amount   string
	Currency string
}

// SearchFilters holds search filter panel state passed from handlers.
type SearchFilters struct {
	Facets        map[string]facet.Filter
	Location      string
	Radius        int
	RadiusUnit    string // "mi" or "km"
	RadiusOptions []int
}

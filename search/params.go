package search

// RadiusMileOptions are valid search radius dropdown values.
var RadiusMileOptions = []int{1, 5, 10, 25, 50, 100}

// Params holds hard filters for listing search.
type Params struct {
	CategoryID int
	Limit      int
	Offset     int
	Q          string
	PriceMin   *int
	PriceMax   *int
	// CenterLat/CenterLon and RadiusKm apply when location text resolved and radius > 0.
	CenterLat float64
	CenterLon float64
	RadiusKm  float64
	HasGeo    bool
}

func (p Params) HasTextQuery() bool {
	return p.Q != ""
}

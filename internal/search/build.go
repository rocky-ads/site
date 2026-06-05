package search

import "github.com/rocky-ads/site/internal/location"

// BuildInput holds filter fields used to construct search Params.
type BuildInput struct {
	CategoryID int
	Limit      int
	Offset     int
	Q          string
	PriceMin   *int
	PriceMax   *int
	Location   string
	Radius     int
	RadiusUnit string
	MileageMin *int
	MileageMax *int
	HoursMin   *int
	HoursMax   *int
}

// BuildParams converts filter input into search Params, resolving geo when applicable.
func BuildParams(in BuildInput) Params {
	p := Params{
		CategoryID: in.CategoryID,
		Limit:      in.Limit,
		Offset:     in.Offset,
		Q:          in.Q,
		PriceMin:   in.PriceMin,
		PriceMax:   in.PriceMax,
		MileageMin: in.MileageMin,
		MileageMax: in.MileageMax,
		HoursMin:   in.HoursMin,
		HoursMax:   in.HoursMax,
	}

	if in.Location == "" || in.Radius <= 0 {
		return p
	}

	lat, lon, ok, err := location.ResolveLocation(in.Location)
	if err != nil || !ok {
		return p
	}

	p.CenterLat = lat
	p.CenterLon = lon
	if in.RadiusUnit == location.UnitKm {
		p.RadiusKm = float64(in.Radius)
	} else {
		p.RadiusKm = location.MilesToKm(float64(in.Radius))
	}
	p.HasGeo = true
	return p
}

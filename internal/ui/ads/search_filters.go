package ads

import (
	"strconv"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/location"
	"github.com/rocky-ads/site/internal/search"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SearchFilters struct {
	PriceMin   *int
	PriceMax   *int
	MileageMin *int
	MileageMax *int
	HoursMin   *int
	HoursMax   *int
	Location   string
	Radius     int
	RadiusUnit string // "mi" or "km"
}

// SearchFiltersPanel renders price, location, and radius controls for the search widget.
func SearchFiltersPanel(category ad.Category, f SearchFilters) g.Node {
	nodes := []g.Node{
		searchPriceRow(f),
	}
	if category.HasMileage() {
		nodes = append(nodes, searchMileageRow(f))
	}
	if category.HasHours() {
		nodes = append(nodes, searchHoursRow(f))
	}
	nodes = append(nodes, searchLocationRadiusRow(f))
	return Div(
		Class("grid grid-cols-2 gap-4"),
		g.Group(nodes),
	)
}

func searchMileageRow(f SearchFilters) g.Node {
	return rangeFilterRow("mileage", "Mileage", f.MileageMin, f.MileageMax)
}

func searchHoursRow(f SearchFilters) g.Node {
	return rangeFilterRow("hours", "Hours", f.HoursMin, f.HoursMax)
}

func rangeFilterRow(name, label string, min, max *int) g.Node {
	minID := "filter-" + name + "-min"
	maxID := "filter-" + name + "-max"
	return Div(
		Class("col-span-2"),
		Label(For(minID), Class("field-label"), g.Text(label)),
		Div(Class("ad-filter-price-range flex flex-wrap items-center gap-2"),
			Input(
				Type("number"),
				Name(name+"_min"),
				ID(minID),
				Class("w-36 p-2 border rounded-md"),
				g.Attr("placeholder", "Min"),
				g.Attr("min", "0"),
				g.Attr("inputmode", "numeric"),
				g.If(priceAmount(min) != "", Value(priceAmount(min))),
			),
			Span(Class("ad-filter-range-sep"), g.Text("–")),
			Input(
				Type("number"),
				Name(name+"_max"),
				ID(maxID),
				Class("w-36 p-2 border rounded-md"),
				g.Attr("placeholder", "Max"),
				g.Attr("min", "0"),
				g.Attr("inputmode", "numeric"),
				g.If(priceAmount(max) != "", Value(priceAmount(max))),
			),
		),
	)
}

func searchPriceRow(f SearchFilters) g.Node {
	return rangeFilterRow("price", "Price", f.PriceMin, f.PriceMax)
}

func searchLocationRadiusRow(f SearchFilters) g.Node {
	unit := f.RadiusUnit
	if unit == "" {
		unit = location.UnitMiles
	}
	radiusLabel := "Radius (miles)"
	if unit == location.UnitKm {
		radiusLabel = "Radius (km)"
	}
	return Div(
		Class("col-span-2 grid grid-cols-2 gap-4"),
		Div(
			Label(For("filter-location"), Class("field-label"), g.Text("Location")),
			LocationInput("filter-location", "location", f.Location, "City or state"),
		),
		Div(
			Label(For("filter-radius"), Class("field-label"), g.Text(radiusLabel)),
			radiusSelect(f.Radius, unit),
		),
	)
}

func radiusSelect(selected int, unit string) g.Node {
	if selected == 0 {
		selected = defaultRadius
	}
	radiusOpts := search.RadiusMileOptions
	suffix := " mi"
	if unit == location.UnitKm {
		radiusOpts = search.RadiusKmOptions
		suffix = " km"
	}
	opts := make([]g.Node, 0, len(radiusOpts))
	for _, n := range radiusOpts {
		label := strconv.Itoa(n) + suffix
		opt := Option(Value(strconv.Itoa(n)), g.Text(label))
		if n == selected {
			opt = Option(Value(strconv.Itoa(n)), g.Attr("selected", "selected"), g.Text(label))
		}
		opts = append(opts, opt)
	}
	return Select(
		Name("radius"),
		ID("filter-radius"),
		Class("w-full p-2 border rounded-md"),
		g.Group(opts),
	)
}

const defaultRadius = 25

func priceAmount(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}

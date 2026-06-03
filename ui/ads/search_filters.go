package ads

import (
	"strconv"

	"github.com/rocky-ads/site/search"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SearchFilters struct {
	PriceMin    *int
	PriceMax    *int
	Location    string
	RadiusMiles int
}

// SearchFiltersPanel renders price, location, and radius controls for the search widget.
func SearchFiltersPanel(f SearchFilters) g.Node {
	return Div(
		Class("grid grid-cols-2 gap-4"),
		searchPriceRow(f),
		searchLocationRadiusRow(f),
	)
}

func searchPriceRow(f SearchFilters) g.Node {
	return Div(
		Class("col-span-2"),
		Label(For("filter-price-min"), Class("field-label"), g.Text("Price")),
		Div(Class("ad-filter-price-range"),
			Input(
				Type("number"),
				Name("price_min"),
				ID("filter-price-min"),
				Class("w-full p-2 border rounded-md"),
				g.Attr("placeholder", "Min"),
				g.Attr("min", "0"),
				g.Attr("inputmode", "numeric"),
				g.If(priceAmount(f.PriceMin) != "", Value(priceAmount(f.PriceMin))),
			),
			Span(Class("ad-filter-range-sep"), g.Text("–")),
			Input(
				Type("number"),
				Name("price_max"),
				ID("filter-price-max"),
				Class("w-full p-2 border rounded-md"),
				g.Attr("placeholder", "Max"),
				g.Attr("min", "0"),
				g.Attr("inputmode", "numeric"),
				g.If(priceAmount(f.PriceMax) != "", Value(priceAmount(f.PriceMax))),
			),
		),
	)
}

func searchLocationRadiusRow(f SearchFilters) g.Node {
	return Div(
		Class("col-span-2 grid grid-cols-2 gap-4"),
		Div(
			Label(For("filter-location"), Class("field-label"), g.Text("Location")),
			LocationInput("filter-location", "location", f.Location, "City or state"),
		),
		Div(
			Label(For("filter-radius"), Class("field-label"), g.Text("Radius (miles)")),
			radiusSelect(f.RadiusMiles),
		),
	)
}

func radiusSelect(selected int) g.Node {
	opts := []g.Node{Option(Value(""), g.Text("Any distance"))}
	for _, miles := range search.RadiusMileOptions {
		opt := Option(Value(strconv.Itoa(miles)), g.Text(strconv.Itoa(miles)+" mi"))
		if miles == selected {
			opt = Option(Value(strconv.Itoa(miles)), g.Attr("selected", "selected"), g.Text(strconv.Itoa(miles)+" mi"))
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

func priceAmount(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}

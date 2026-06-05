package ads

import (
	"fmt"
	"strconv"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/location"
	"github.com/rocky-ads/site/internal/search"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SearchFilters struct {
	Facets     map[string]facet.Filter
	Location   string
	Radius     int
	RadiusUnit string // "mi" or "km"
}

// SearchFiltersPanel renders facet, location, and radius controls for the search widget.
func SearchFiltersPanel(category ad.Category, f SearchFilters) g.Node {
	var nodes []g.Node
	for _, d := range category.Facets() {
		if !d.Filterable {
			continue
		}
		nodes = append(nodes, facetFilterRow(d, f.Facets[d.Key]))
	}
	nodes = append(nodes, searchLocationRadiusRow(f))
	return Div(
		Class("grid grid-cols-2 gap-4"),
		g.Group(nodes),
	)
}

func facetFilterRow(d facet.Def, filter facet.Filter) g.Node {
	switch d.Filter {
	case facet.FilterExact:
		return enumFilterRow(d, filter)
	case facet.FilterCheckboxes:
		return enumCheckboxesFilterRow(d, filter)
	default:
		return rangeFilterRow(d.Key, d.Label, filter.Min, filter.Max)
	}
}

func enumFilterRow(d facet.Def, filter facet.Filter) g.Node {
	selected := ""
	if filter.Value != nil {
		selected = *filter.Value
	}
	id := "filter-" + d.Key
	opts := make([]g.Node, 0, len(d.Enum)+1)
	opts = append(opts, enumOption("", "Any", selected))
	for _, e := range d.Enum {
		opts = append(opts, enumOption(e, e, selected))
	}
	return Div(
		Class("col-span-2"),
		Label(For(id), Class("field-label"), g.Text(d.Label)),
		Select(
			Name(d.Key),
			ID(id),
			Class("w-full p-2 border rounded-md"),
			g.Group(opts),
		),
	)
}

func enumCheckboxesFilterRow(d facet.Def, filter facet.Filter) g.Node {
	selected := make(map[string]bool, len(filter.Values))
	for _, v := range filter.Values {
		selected[v] = true
	}
	nodes := make([]g.Node, len(d.Enum))
	for i, e := range d.Enum {
		id := fmt.Sprintf("filter-%s-%d", d.Key, i)
		attrs := []g.Node{
			Type("checkbox"),
			Name(d.Key),
			Value(e),
			ID(id),
		}
		if selected[e] {
			attrs = append(attrs, g.Attr("checked", "checked"))
		}
		nodes[i] = Label(
			Class("flex items-center gap-2"),
			Input(attrs...),
			g.Text(e),
		)
	}
	return Div(
		Class("col-span-2"),
		Label(Class("field-label"), g.Text(d.Label)),
		Div(Class("flex flex-wrap items-center gap-4"), g.Group(nodes)),
	)
}

func enumOption(value, label, selected string) g.Node {
	if value == selected {
		return Option(Value(value), g.Attr("selected", "selected"), g.Text(label))
	}
	return Option(Value(value), g.Text(label))
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

package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldSelect renders a select dropdown field with an "All" option
func FieldSelect(name, displayName, selectedValue string, values []string) g.Node {

	var options []g.Node
	options = append(options, Option(Value(""), g.Text("All"), g.If(selectedValue == "", Selected())))
	for _, val := range values {
		options = append(options, Option(Value(val), g.Text(val), g.If(val == selectedValue, Selected())))
	}

	return Div(
		Label(Class("block text-sm font-medium mb-1"), g.Text(displayName)),
		Select(
			Name(name),
			Class("w-full p-2 border rounded-md"),
			g.Group(options),
		),
	)
}

func PriceRange(displayName, minPrice, maxPrice string) g.Node {
	return Div(
		Label(Class("block text-sm font-medium mb-1"), g.Text(displayName+" Range")),
		Div(
			Class("flex gap-2 flex-nowrap"),
			Input(
				Type("number"),
				Name("min_price"),
				ID("minPriceFilter"),
				Class("w-24 flex-shrink-0 p-2 border rounded-md"),
				Placeholder("Min $"),
				Min("0"),
				Step("1"),
				Value(minPrice),
			),
			Input(
				Type("number"),
				Name("max_price"),
				ID("maxPriceFilter"),
				Class("w-24 flex-shrink-0 p-2 border rounded-md"),
				Placeholder("Max $"),
				Min("0"),
				Step("1"),
				Value(maxPrice),
			),
		),
	)
}

func YearRange(displayName, minYear, maxYear, maxMaxYear string) g.Node {
	return Div(
		Label(Class("block text-sm font-medium mb-1"), g.Text(displayName+" Range")),
		Div(
			Class("flex gap-2 flex-nowrap"),
			Input(
				Type("number"),
				Name("min_year"),
				ID("minYearFilter"),
				Class("w-20 flex-shrink-0 p-2 border rounded-md"),
				Placeholder("Min"),
				Min("1900"),
				Max(maxMaxYear),
				Value(minYear),
			),
			Input(
				Type("number"),
				Name("max_year"),
				ID("maxYearFilter"),
				Class("w-20 flex-shrink-0 p-2 border rounded-md"),
				Placeholder("Max"),
				Min("1900"),
				Max(maxMaxYear),
				Value(maxYear),
			),
		),
	)
}

func LocationRadius(location, radius string) g.Node {
	return g.Group([]g.Node{
		// Location input
		Div(
			Label(Class("block text-sm font-medium mb-1"), g.Text("Location")),
			Input(
				Type("text"),
				Name("location"),
				Class("w-full p-2 border rounded-md"),
				Placeholder("City, State or ZIP"),
				Value(location),
			),
		),
		// Radius dropdown
		Div(
			Label(Class("block text-sm font-medium mb-1"), g.Text("Radius")),
			Select(
				Name("radius"),
				Class("w-full p-2 border rounded-md"),
				Option(Value("25"), g.Text("25 miles"), g.If(radius == "25", Selected())),
				Option(Value("50"), g.Text("50 miles"), g.If(radius == "50", Selected())),
				Option(Value("100"), g.Text("100 miles"), g.If(radius == "100", Selected())),
				Option(Value("250"), g.Text("250 miles"), g.If(radius == "250", Selected())),
				Option(Value("500"), g.Text("500 miles"), g.If(radius == "500", Selected())),
			),
		),
	})
}

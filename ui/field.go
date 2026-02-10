package ui

import (
	"fmt"
	"time"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldSelect renders a select dropdown. defaultName is the label for the empty value (e.g. "All" for filters, "Select a Make" for new ad).
func FieldSelect(name, displayName, defaultName, selectedValue string, values []string) g.Node {

	var options []g.Node
	options = append(options, Option(Value(""), g.Text(defaultName), g.If(selectedValue == "", Selected())))
	for _, val := range values {
		options = append(options, Option(Value(val), g.Text(val), g.If(val == selectedValue, Selected())))
	}

	return Div(
		label(displayName),
		Select(
			Name(name),
			Class("w-full p-2 border rounded-md"),
			g.Group(options),
		),
	)
}

func PriceRange(displayName, minPrice, maxPrice string) g.Node {
	return Div(
		label(displayName+" Range"),
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
		label(displayName+" Range"),
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
			label("Location"),
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
			label("Radius"),
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

func LocationInput(isRequired bool) g.Node {
	return Div(
		label("Location"),
		inputText("location", "City, State or ZIP (optional)", isRequired,
			MaxLength("32"),
			Pattern("[\\x20-\\x7E]+"),
			g.Attr("oninput", "this.checkValidity()"),
		),
	)
}

func PriceInput(isRequired bool) g.Node {
	return Div(
		label("Price"),
		inputText("price", "0", isRequired,
			Type("number"),
			Min("0"),
			Step("1"),
			Pattern("^(0|[1-9][0-9]*)$"),
			Title("Price must be a non-negative integer"),
			g.Attr("oninput", "this.checkValidity()"),
		),
	)
}

// YearInput renders a single year input field for forms
func YearInput(displayName, value string, isRequired bool) g.Node {
	maxYear := time.Now().Year() + 2
	inputAttrs := []g.Node{
		Type("number"),
		Name("year"),
		Class("w-full p-2 border rounded-md"),
		Placeholder("2024"),
		Min("1900"),
		Max(fmt.Sprintf("%d", maxYear)),
		Step("1"),
		Value(value),
	}
	if isRequired {
		inputAttrs = append(inputAttrs, Required())
	}

	return Div(
		label(displayName),
		Input(inputAttrs...),
	)
}

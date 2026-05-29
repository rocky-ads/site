package ui

import (
	"strings"

	"github.com/rocky-ads/site/currency"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// fieldURL builds the htmx GET URL for the next field, baking in previous query params.
func fieldURL(nextField, prevParams string) string {
	url := "/api/field/" + nextField
	if prevParams != "" {
		url += "?" + prevParams
	}
	return url
}

// FieldSelect renders a select dropdown.
func FieldSelect(name, displayName, selectedValue, nextField, prevParams string, values []string, required bool) g.Node {

	defaultOptionText := "All"
	if required {
		defaultOptionText = "Select a " + displayName
	}
	var options []g.Node
	options = append(options, Option(Value(""), g.Text(defaultOptionText), g.If(selectedValue == "", Selected())))
	for _, val := range values {
		options = append(options, Option(Value(val), g.Text(val), g.If(val == selectedValue, Selected())))
	}

	selectAttrs := []g.Node{
		Name(name),
		Class("w-full p-2 border rounded-md"),
		g.If(required, Required()),
		g.Group(options),
	}
	if nextField != "" {
		selectAttrs = append(selectAttrs,
			hx.Get(fieldURL(nextField, prevParams)),
			hx.Trigger("change"),
			hx.Target("#div-"+nextField),
			hx.Swap("outerHTML"),
		)
	}

	return Div(
		Class("mt-3"),
		label(displayName),
		Select(selectAttrs...),
		g.If(nextField != "", FieldFragment(nextField, g.Group{})),
	)
}

// FieldFragment wraps a field node in a div with id "div-"+name for htmx swap targets.
func FieldFragment(name string, content g.Node) g.Node {
	return Div(ID("div-"+name), content)
}

// FieldCheckboxes renders a grid of checkboxes for the given values.
func FieldCheckboxes(name, displayName, nextField, prevParams string, values []string) g.Node {
	boxClass := "border rounded-md p-3"

	if len(values) == 0 {
		emptyDiv := Div(
			Class(boxClass+" text-sm text-zinc-400 italic"),
			g.Textf("No %ss match the selected filters", strings.ToLower(displayName)),
		)
		return Div(
			Class("mt-3"),
			label(displayName),
			ID(name),
			emptyDiv,
		)
	}

	var nodes []g.Node
	for _, val := range values {
		nodes = append(nodes, checkbox(name, val, val, false, false))
	}
	gridClass := "grid grid-cols-4 sm:grid-cols-6 gap-2 " + boxClass
	gridContent := g.Group(nodes)
	gridDiv := Div(Class(gridClass), gridContent)
	if nextField != "" {
		gridDiv = Div(
			Class(gridClass),
			hx.Get(fieldURL(nextField, prevParams)),
			hx.Trigger("change from:input"),
			hx.Target("#div-"+nextField),
			hx.Include("this"),
			hx.Swap("outerHTML"),
			gridContent,
		)
	}
	return Div(
		Class("mt-3"),
		label(displayName),
		ID(name),
		gridDiv,
		g.If(nextField != "", FieldFragment(nextField, g.Group{})),
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
	locationLabel := "Location"
	if !isRequired {
		locationLabel = "Location (optional)"
	}
	return Div(
		label(locationLabel),
		inputText("location", "City, State or ZIP", isRequired,
			MaxLength("32"),
			Pattern("[\\x20-\\x7E]+"),
			g.Attr("oninput", "this.checkValidity()"),
		),
	)
}

func PriceInput(isRequired bool, defaultCurrency string) g.Node {
	defaultCurrency = currency.Normalize(defaultCurrency)
	if !currency.IsSupported(defaultCurrency) {
		defaultCurrency = currency.Default
	}
	opts := make([]g.Node, len(currency.Supported))
	for i, code := range currency.Supported {
		opt := Option(Value(code), g.Text(code))
		if code == defaultCurrency {
			opt = Option(Value(code), g.Attr("selected", "selected"), g.Text(code))
		}
		opts[i] = opt
	}
	return Div(
		label("Price"),
		Div(
			Class("flex gap-2 flex-nowrap"),
			inputText("price", "0", isRequired,
				Type("number"),
				Min("0"),
				Step("1"),
				Pattern("^(0|[1-9][0-9]*)$"),
				Title("Price must be a non-negative integer"),
				g.Attr("oninput", "this.checkValidity()"),
				Class("flex-1 p-2 border rounded-md"),
			),
			Select(
				Name("price_currency"),
				ID("price-currency"),
				Class("p-2 border rounded-md"),
				g.Group(opts),
			),
		),
		P(Class("text-sm text-zinc-500 mt-1"), g.Text("Enter 0 for FREE.")),
	)
}

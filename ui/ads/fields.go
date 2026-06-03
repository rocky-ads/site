package ads

import (
	"strconv"

	"github.com/rocky-ads/site/currency"
	"github.com/rocky-ads/site/validation"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// NewAdFieldsPartial renders static new-ad fields: title, description, location, price.
func NewAdFieldsPartial(defaultCurrency string) g.Node {
	titleSpec := validation.ParseInputType("text?max=35&pattern=ascii")
	descSpec := validation.ParseInputType("text?max=1000&pattern=ascii-multiline")
	currencyCode := priceCurrencyCode("", defaultCurrency)

	return Div(
		ID("category-fields"),
		Class("category-fields space-y-6"),
		fieldBlock("Title", titleInput(titleSpec)),
		fieldBlock("Description", descriptionInput(descSpec)),
		fieldBlock("Location (optional)", LocationInput("new-ad-location", "location", "", "City or state")),
		fieldBlock("Price", newAdPriceRow(currencyCode)),
	)
}

func fieldBlock(label string, input g.Node) g.Node {
	return Div(Class("mt-3"), Label(Class("field-label"), g.Text(label)), input)
}

func titleInput(spec validation.InputSpec) g.Node {
	attrs := []g.Node{
		Type("text"),
		Name("title"),
		ID("new-ad-title"),
		Class("w-full p-2 border rounded-md"),
		g.Attr("required", "required"),
	}
	attrs = appendTextValidation(attrs, spec)
	return Input(attrs...)
}

func descriptionInput(spec validation.InputSpec) g.Node {
	attrs := []g.Node{
		Name("description"),
		ID("new-ad-description"),
		Class("w-full p-2 border rounded-md"),
		g.Attr("rows", "6"),
		g.Attr("required", "required"),
	}
	attrs = appendTextareaValidation(attrs, spec)
	return Textarea(attrs...)
}

func newAdPriceRow(currencyCode string) g.Node {
	return Div(
		Class("flex flex-wrap items-center gap-2"),
		Input(
			Type("number"),
			Name("price"),
			ID("new-ad-price"),
			Class("w-36 p-2 border rounded-md"),
			g.Attr("min", "0"),
			g.Attr("step", "1"),
			g.Attr("inputmode", "numeric"),
		),
		priceCurrencySelectStatic(currencyCode),
		Label(
			Class("flex items-center gap-2"),
			Input(Type("checkbox"), Name("price_free"), Value("1"), ID("new-ad-price-free")),
			Span(g.Text("List as FREE")),
		),
	)
}

func priceCurrencySelectStatic(selected string) g.Node {
	selected = currency.Normalize(selected)
	if !currency.IsSupported(selected) {
		selected = currency.Default
	}
	opts := make([]g.Node, len(currency.Supported))
	for i, code := range currency.Supported {
		opt := Option(Value(code), g.Text(code))
		if code == selected {
			opt = Option(Value(code), g.Attr("selected", "selected"), g.Text(code))
		}
		opts[i] = opt
	}
	return Select(
		Name("price_currency"),
		ID("new-ad-price-currency"),
		Class("w-24 p-2 border rounded-md shrink-0"),
		g.Group(opts),
	)
}

func priceCurrencyCode(selected, defaultCurrency string) string {
	code := currency.Normalize(selected)
	if code == "" || !currency.IsSupported(code) {
		code = currency.Normalize(defaultCurrency)
	}
	if !currency.IsSupported(code) {
		return currency.Default
	}
	return code
}

func appendTextValidation(attrs []g.Node, spec validation.InputSpec) []g.Node {
	if max := spec.Param("max"); max != "" {
		attrs = append(attrs, g.Attr("maxlength", max))
	}
	if pat := spec.HTMLPattern(); pat != "" {
		attrs = append(attrs, g.Attr("pattern", pat))
		if msg := spec.PatternMessage(); msg != "" {
			quoted := strconv.Quote(msg)
			attrs = append(attrs, g.Attr("title", msg), g.Attr("oninvalid", "this.setCustomValidity("+quoted+")"), g.Attr("oninput", "this.setCustomValidity('')"))
		}
	}
	return attrs
}

func appendTextareaValidation(attrs []g.Node, spec validation.InputSpec) []g.Node {
	if max := spec.Param("max"); max != "" {
		attrs = append(attrs, g.Attr("maxlength", max))
	}
	if pat := spec.HTMLPattern(); pat != "" {
		msg := spec.PatternMessage()
		anchored := validation.AnchoredPattern(pat)
		check := `(function(el){var ok=new RegExp(` + strconv.Quote(anchored) + `).test(el.value);el.setCustomValidity(ok?'':` + strconv.Quote(msg) + `);}).call(null,this)`
		attrs = append(attrs, g.Attr("title", msg), g.Attr("oninput", check), g.Attr("onchange", check))
	}
	return attrs
}

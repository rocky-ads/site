package ads

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/facet"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

const newAdPriceRowID = "new-ad-price-row"

const (
	asciiPattern          = `[\x20-\x7E]+`
	asciiMultilinePattern = `[\x20-\x7E\n\r]+`
	asciiMsg              = "Please enter printable ASCII characters only"
	asciiMultilineMsg     = "Please enter printable ASCII characters only (line breaks allowed)"
)

type adFields struct {
	defaults facet.FormDefaults
}

// NewAdFieldsPartial renders new-ad fields for the given category facets.
func NewAdFieldsPartial(facets []facet.Def, defaults facet.FormDefaults) g.Node {
	f := adFields{defaults: defaults}
	nodes := []g.Node{
		fieldBlock("Title", titleInput()),
		fieldBlock("Description", descriptionInput()),
		fieldBlock("Location (optional)", LocationInput("new-ad-location", "location", "", "City or state")),
	}
	for _, d := range facets {
		nodes = append(nodes, facetFieldBlock(d, f.facetInput(d)))
	}

	return Div(
		ID("category-fields"),
		Class("category-fields space-y-6"),
		g.Group(nodes),
	)
}

func (f adFields) facetInput(d facet.Def) g.Node {
	switch d.Form {
	case facet.FormMoney:
		return f.newAdPriceRow(d)
	case facet.FormSelect:
		return formSelect(d)
	case facet.FormRadio:
		return formRadio(d)
	default:
		if len(d.Units) > 0 {
			return f.intWithUnitRow(d)
		}
		return facetNumberInput(d)
	}
}

func formSelect(d facet.Def) g.Node {
	opts := make([]g.Node, 0, len(d.FormOptions())+1)
	opts = append(opts, Option(Value(""), g.Text(selectPlaceholder(d))))
	for _, o := range d.FormOptions() {
		opts = append(opts, Option(Value(o), g.Text(o)))
	}
	attrs := []g.Node{
		Name(d.Key),
		ID("new-ad-" + d.Key),
		Class("w-full p-2 border rounded-md"),
	}
	attrs = append(attrs, requiredAttr(d.Required)...)
	attrs = append(attrs, g.Group(opts))
	return Select(attrs...)
}

func selectPlaceholder(d facet.Def) string {
	return "Select a " + d.Key + "..."
}

func formRadio(d facet.Def) g.Node {
	opts := d.FormOptions()
	nodes := make([]g.Node, len(opts))
	for i, o := range opts {
		id := fmt.Sprintf("new-ad-%s-%d", d.Key, i)
		attrs := []g.Node{
			Type("radio"),
			Name(d.Key),
			Value(o),
			ID(id),
		}
		if d.Required && i == 0 {
			attrs = append(attrs, g.Attr("required", "required"))
		}
		nodes[i] = Label(
			Class("flex items-center gap-2"),
			Input(attrs...),
			Span(g.Text(o)),
		)
	}
	return Div(Class("flex flex-wrap items-center gap-4"), g.Group(nodes))
}

func facetNumberInput(d facet.Def) g.Node {
	attrs := []g.Node{
		Type("number"),
		Name(d.Key),
		ID("new-ad-" + d.Key),
		Class("w-full p-2 border rounded-md"),
		g.Attr("min", "0"),
		g.Attr("step", "1"),
		g.Attr("inputmode", "numeric"),
	}
	attrs = append(attrs, requiredAttr(d.Required)...)
	return Input(attrs...)
}

func (f adFields) intWithUnitRow(d facet.Def) g.Node {
	selected := d.FormDefaultUnit(f.defaults)
	unitName := d.Key + "_unit"
	numAttrs := []g.Node{
		Type("number"),
		Name(d.Key),
		ID("new-ad-" + d.Key),
		Class("w-36 p-2 border rounded-md"),
		g.Attr("min", "0"),
		g.Attr("step", "1"),
		g.Attr("inputmode", "numeric"),
	}
	numAttrs = append(numAttrs, requiredAttr(d.Required)...)
	return Div(
		Class("flex flex-wrap items-center gap-2"),
		Input(numAttrs...),
		facetUnitSelect(unitName, selected, d.Units),
	)
}

func facetUnitSelect(name, selected string, units []string) g.Node {
	opts := make([]g.Node, len(units))
	for i, u := range units {
		opt := Option(Value(u), g.Text(u))
		if u == selected {
			opt = Option(Value(u), g.Attr("selected", "selected"), g.Text(u))
		}
		opts[i] = opt
	}
	return Select(
		Name(name),
		ID("new-ad-"+name),
		Class("p-2 border rounded-md shrink-0"),
		g.Group(opts),
	)
}

func facetFieldBlock(d facet.Def, input g.Node) g.Node {
	label := d.Label
	if !d.Required {
		label += " (optional)"
	}
	return fieldBlock(label, input)
}

func fieldBlock(label string, input g.Node) g.Node {
	return Div(Class("mt-3"), Label(Class("field-label"), g.Text(label)), input)
}

func requiredAttr(required bool) []g.Node {
	if !required {
		return nil
	}
	return []g.Node{g.Attr("required", "required")}
}

func titleInput() g.Node {
	max := strconv.Itoa(config.MaxAdTitleLength)
	quotedMsg := strconv.Quote(asciiMsg)
	return Input(
		Type("text"),
		Name("title"),
		ID("new-ad-title"),
		Class("w-full p-2 border rounded-md"),
		g.Attr("required", "required"),
		g.Attr("maxlength", max),
		g.Attr("pattern", asciiPattern),
		g.Attr("title", asciiMsg),
		g.Attr("oninvalid", "this.setCustomValidity("+quotedMsg+")"),
		g.Attr("oninput", "this.setCustomValidity('')"),
	)
}

func descriptionInput() g.Node {
	max := strconv.Itoa(config.MaxAdDescriptionLength)
	anchored := "^(?:" + asciiMultilinePattern + ")$"
	check := `(function(el){var ok=new RegExp(` + strconv.Quote(anchored) + `).test(el.value);el.setCustomValidity(ok?'':` + strconv.Quote(asciiMultilineMsg) + `);}).call(null,this)`
	return Textarea(
		Name("description"),
		ID("new-ad-description"),
		Class("w-full p-2 border rounded-md"),
		g.Attr("rows", "6"),
		g.Attr("required", "required"),
		g.Attr("maxlength", max),
		g.Attr("data-pattern-check", ""),
		g.Attr("title", asciiMultilineMsg),
		g.Attr("oninput", check),
		g.Attr("onchange", check),
	)
}

func (f adFields) newAdPriceRow(d facet.Def) g.Node {
	return NewAdPriceRow(d, f.defaults, PriceRowView{
		Currency: d.FormDefaultCurrency(f.defaults),
	})
}

// NewAdPriceRow renders the price facet with a "List as FREE" HTMX toggle.
func NewAdPriceRow(d facet.Def, defaults facet.FormDefaults, view PriceRowView) g.Node {
	currencyCode := priceCurrencyCode(d, defaults, view.Currency)

	freeCheckbox := Label(
		Class("flex items-center gap-2 mb-2"),
		Input(priceFreeCheckboxAttrs(view.IsFree)...),
		Span(g.Text("List as FREE")),
	)

	var body g.Node
	if view.IsFree {
		body = Div(
			Class("price-free-state"),
			Input(Type("hidden"), Name(d.Key), Value("0")),
			Input(Type("hidden"), Name("price_currency"), Value(currencyCode)),
		)
	} else {
		amountAttrs := []g.Node{
			Type("number"),
			Name(d.Key),
			ID("new-ad-" + d.Key),
			Class("w-36 p-2 border rounded-md"),
			g.Attr("min", "0"),
			g.Attr("step", "1"),
			g.Attr("inputmode", "numeric"),
		}
		if amount := strings.TrimSpace(view.Amount); amount != "" && amount != "0" {
			amountAttrs = append(amountAttrs, Value(amount))
		}
		body = Div(
			Class("price-priced-state"),
			Div(
				Class("flex flex-wrap items-center gap-2"),
				Input(amountAttrs...),
				priceCurrencySelect(currencyCode, d.SupportedCurrencies()),
			),
		)
	}

	return Div(
		ID(newAdPriceRowID),
		freeCheckbox,
		body,
	)
}

func priceFreeCheckboxAttrs(checked bool) []g.Node {
	attrs := []g.Node{
		Type("checkbox"),
		Name("price_free"),
		Value("1"),
		ID("new-ad-price-free"),
		hx.Get("/auth/ad/new/price-field"),
		hx.Target("#" + newAdPriceRowID),
		hx.Swap("outerHTML"),
		hx.Trigger("change"),
		hx.Include("#" + newAdPriceRowID),
	}
	if checked {
		attrs = append(attrs, g.Attr("checked", "checked"))
	}
	return attrs
}

func priceCurrencyCode(d facet.Def, defaults facet.FormDefaults, selected string) string {
	if selected != "" {
		return d.FormDefaultCurrency(facet.FormDefaults{Currency: selected})
	}
	return d.FormDefaultCurrency(defaults)
}

func priceCurrencySelect(selected string, currencies []string) g.Node {
	opts := make([]g.Node, len(currencies))
	for i, code := range currencies {
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

package ads

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rocky-ads/site/currency"
	"github.com/rocky-ads/site/field"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
	hx "maragu.dev/gomponents-htmx"
)

type FieldsView struct {
	Category        field.CategoryChains
	OptionsMap      map[int]map[string][]string // chainID -> fieldName -> options
	DefaultCurrency string
}

func CategoryFieldsPartial(view FieldsView) g.Node {
	if len(view.Category.Chains) == 0 {
		return Div(Class("category-fields"), P(g.Text("No fields configured for this category.")))
	}

	sections := make([]g.Node, 0, len(view.Category.Chains))
	for _, chain := range view.Category.Chains {
		sections = append(sections, chainSection(view.Category.CategoryID, chain, view.OptionsMap[chain.ChainID], view.DefaultCurrency))
	}

	return Div(
		ID("category-fields"),
		Class("category-fields space-y-6"),
		g.Group(sections),
	)
}

func chainSection(categoryID int, chain field.ChainGroup, chainOpts map[string][]string, defaultCurrency string) g.Node {
	fields := make([]g.Node, 0, len(chain.Fields))

	if field.IsSpecChain(chain) {
		if len(chain.Fields) > 0 {
			first := chain.Fields[0]
			var opts []string
			if chainOpts != nil {
				opts = chainOpts[first.FieldName]
			}
			fields = append(fields, specFieldRow(categoryID, chain, first, opts))
			fields = append(fields, chainNextPlaceholder(chain.ChainID, first.FieldName))
		}
	} else {
		for _, f := range chain.Fields {
			fields = append(fields, plainFieldRow(categoryID, chain, f, defaultCurrency))
		}
	}

	section := []g.Node{
		ID(chainContainerID(chain.ChainID)),
		Class("chain-section"),
	}
	if len(chain.Fields) == 1 {
		section = append(section, H3(Class("chain-title"), g.Text(singleFieldChainTitle(chain.Fields[0]))))
	}
	section = append(section, Div(Class("chain-fields"), g.Group(fields)))
	return Section(section...)
}

// NextFieldPartial swaps into #chain-{id}-next-{afterField} and shows one field plus a nested placeholder.
func NextFieldPartial(categoryID int, chain field.ChainGroup, afterField, fieldName string, options []string, noMatch bool) g.Node {
	if fieldName == "" {
		return chainNextPlaceholder(chain.ChainID, afterField)
	}

	if noMatch || len(options) == 0 {
		return chainNoMatchPartial(chain, afterField, fieldName)
	}

	f, ok := findChainField(chain, fieldName)
	if !ok {
		return chainNextPlaceholder(chain.ChainID, afterField)
	}

	nextAfter := nextChainFieldName(chain, fieldName)
	children := []g.Node{specFieldRow(categoryID, chain, f, options)}
	if nextAfter != "" {
		children = append(children, chainNextPlaceholder(chain.ChainID, f.FieldName))
	}

	return Div(
		ID(chainNextID(chain.ChainID, afterField)),
		Class("chain-next"),
		g.Group(children),
	)
}

func chainNoMatchPartial(chain field.ChainGroup, afterField, fieldName string) g.Node {
	child, okChild := findChainField(chain, fieldName)
	parent, okParent := findChainField(chain, afterField)
	childLabel := fieldName
	parentLabel := afterField
	if okChild {
		childLabel = pluralDisplayName(child.DisplayName)
	}
	if okParent {
		parentLabel = parent.DisplayName
	}
	msg := fmt.Sprintf("No %s match the %s selection.", childLabel, parentLabel)

	return Div(
		ID(chainNextID(chain.ChainID, afterField)),
		Class("chain-next"),
		P(Class("field-no-match"), g.Text(msg)),
	)
}

func pluralDisplayName(displayName string) string {
	if strings.HasSuffix(displayName, "s") {
		return displayName
	}
	if strings.HasSuffix(displayName, "y") {
		return displayName[:len(displayName)-1] + "ies"
	}
	return displayName + "s"
}

func chainNextPlaceholder(chainID int, afterField string) g.Node {
	return Div(ID(chainNextID(chainID, afterField)), Class("chain-next"))
}

func specFieldRow(categoryID int, chain field.ChainGroup, f field.ChainField, options []string) g.Node {
	label := g.Node(fieldLabel(f, "select"))
	if field.IsMultiInput(f.InputType) {
		label = P(Class("field-label"), g.Text(fieldLabelText(f)))
	}
	return Div(
		ID(fieldContainerID(chain.ChainID, f.FieldName)),
		Class("mt-3"),
		label,
		specFieldInput(categoryID, chain, f, options),
	)
}

func plainFieldRow(categoryID int, chain field.ChainGroup, f field.ChainField, defaultCurrency string) g.Node {
	id := fieldContainerID(chain.ChainID, f.FieldName)
	spec := field.ParseInputType(f.InputType)
	if f.FieldName == "price" {
		return PriceFieldPartial(categoryID, chain, f, PriceFieldView{
			DefaultCurrency: defaultCurrency,
			IsFree:          false,
		})
	}
	input := plainFieldInput(f, id, spec)

	// Single-field chains use display name as the section title (chainTitle).
	if len(chain.Fields) == 1 {
		return Div(ID(id), Class("mt-3"), input)
	}

	return Div(ID(id), Class("mt-3"), fieldLabel(f, "input-"+id), input)
}

func plainFieldInput(f field.ChainField, id string, spec field.InputSpec) g.Node {
	switch spec.Type {
	case field.InputNumber:
		return numberInput(f, id, spec)
	default:
		if f.FieldName == "description" {
			return textareaInput(f, id, spec)
		}
		return textInput(f, id, spec)
	}
}

func appendValidationAttrs(attrs []g.Node, f field.ChainField, spec field.InputSpec) []g.Node {
	return appendValidationAttrsFor(attrs, f, spec, false)
}

func appendTextareaValidationAttrs(attrs []g.Node, f field.ChainField, spec field.InputSpec) []g.Node {
	return appendValidationAttrsFor(attrs, f, spec, true)
}

func appendValidationAttrsFor(attrs []g.Node, f field.ChainField, spec field.InputSpec, textarea bool) []g.Node {
	if f.IsRequired {
		attrs = append(attrs, g.Attr("required", "required"))
	}
	if max := spec.Param("max"); max != "" {
		attrs = append(attrs, g.Attr("maxlength", max))
	}
	pat := spec.HTMLPattern()
	if pat == "" {
		return attrs
	}
	msg := spec.PatternMessage()
	if textarea {
		// HTML does not apply pattern on <textarea>; validate via Constraint Validation API.
		return append(attrs, textareaPatternAttrs(pat, msg)...)
	}
	attrs = append(attrs, g.Attr("pattern", pat))
	return append(attrs, patternValidityAttrs(msg)...)
}

// patternValidityAttrs sets a custom constraint message (oninvalid/oninput) and title fallback.
func patternValidityAttrs(msg string) []g.Node {
	if msg == "" {
		return nil
	}
	quoted := strconv.Quote(msg)
	return []g.Node{
		g.Attr("title", msg),
		g.Attr("oninvalid", "this.setCustomValidity("+quoted+")"),
		g.Attr("oninput", "this.setCustomValidity('')"),
	}
}

// textareaPatternAttrs checks the regex on input/change because textarea ignores pattern=.
func textareaPatternAttrs(regex, msg string) []g.Node {
	if regex == "" {
		return nil
	}
	anchored := field.AnchoredPattern(regex)
	check := fmt.Sprintf(
		"(function(el){var ok=new RegExp(%s).test(el.value);el.setCustomValidity(ok?'':%s);}).call(null,this)",
		strconv.Quote(anchored), strconv.Quote(msg),
	)
	return []g.Node{
		g.Attr("data-pattern-check", "1"),
		g.Attr("title", msg),
		g.Attr("oninput", check),
		g.Attr("onchange", check),
	}
}

func specFieldInput(categoryID int, chain field.ChainGroup, f field.ChainField, options []string) g.Node {
	if field.IsMultiInput(f.InputType) {
		return specMultiCheckboxGrid(categoryID, chain, f, options)
	}
	return specSelect(categoryID, chain, f, options)
}

func findChainField(chain field.ChainGroup, fieldName string) (field.ChainField, bool) {
	for _, f := range chain.Fields {
		if f.FieldName == fieldName {
			return f, true
		}
	}
	return field.ChainField{}, false
}

func nextChainFieldName(chain field.ChainGroup, current string) string {
	found := false
	for _, f := range chain.Fields {
		if found {
			return f.FieldName
		}
		if f.FieldName == current {
			found = true
		}
	}
	return ""
}

func fieldLabelText(f field.ChainField) string {
	if f.IsRequired {
		return f.DisplayName
	}
	return f.DisplayName + " (optional)"
}

func singleFieldChainTitle(f field.ChainField) string {
	return fieldLabelText(f)
}

func numberInput(f field.ChainField, id string, spec field.InputSpec) g.Node {
	inputID := "input-" + id
	attrs := []g.Node{
		Name(f.FieldName),
		ID(inputID),
		Class("w-full p-2 border rounded-md"),
	}
	if spec.HTMLPattern() != "" {
		attrs = append([]g.Node{Type("text"), g.Attr("inputmode", "numeric")}, attrs...)
	} else {
		attrs = append([]g.Node{Type("number")}, attrs...)
		if min := spec.Param("min"); min != "" {
			attrs = append(attrs, g.Attr("min", min))
		}
	}
	return Input(appendValidationAttrs(attrs, f, spec)...)
}

// PriceFieldView holds state for the price row partial (initial render and HTMX swaps).
type PriceFieldView struct {
	DefaultCurrency string
	IsFree          bool
	Amount          string
	Currency        string
}

// PriceFieldPartial renders the price field with a "List as FREE" HTMX toggle.
func PriceFieldPartial(categoryID int, chain field.ChainGroup, f field.ChainField, view PriceFieldView) g.Node {
	id := fieldContainerID(chain.ChainID, f.FieldName)
	inputID := "input-" + id
	currencyCode := priceCurrencyCode(view.Currency, view.DefaultCurrency)

	freeCheckbox := Label(
		Class("flex items-center gap-2 mb-2"),
		Input(priceFreeCheckboxAttrs(categoryID, chain.ChainID, id, view.IsFree)...),
		Span(g.Text("List as FREE")),
	)

	var body g.Node
	if view.IsFree {
		body = Div(
			Class("price-free-state"),
			Input(Type("hidden"), Name(f.FieldName), Value("0")),
			Input(Type("hidden"), Name("price_currency"), Value(currencyCode)),
		)
	} else {
		amountAttrs := []g.Node{
			Type("number"),
			Name(f.FieldName),
			ID(inputID),
			Class("w-36 p-2 border rounded-md"),
			g.Attr("min", "1"),
			g.Attr("step", "1"),
			g.Attr("inputmode", "numeric"),
		}
		if amount := strings.TrimSpace(view.Amount); amount != "" && amount != "0" {
			amountAttrs = append(amountAttrs, Value(amount))
		}
		if f.IsRequired {
			amountAttrs = append(amountAttrs, g.Attr("required", "required"))
		}
		body = Div(
			Class("price-priced-state"),
			Div(Class("flex flex-wrap items-center gap-2"),
				Input(amountAttrs...),
				priceCurrencySelect(currencyCode),
			),
		)
	}

	content := []g.Node{freeCheckbox, body}
	if len(chain.Fields) == 1 {
		return Div(ID(id), Class("mt-3"), g.Group(content))
	}
	return Div(
		ID(id),
		Class("mt-3"),
		fieldLabel(f, inputID),
		g.Group(content),
	)
}

func priceFreeCheckboxAttrs(categoryID, chainID int, containerID string, checked bool) []g.Node {
	// name/value only sent when checked; handler treats missing price_free as priced.
	attrs := []g.Node{
		Type("checkbox"),
		Name("price_free"),
		Value("1"),
		hx.Get(priceFieldURL(categoryID, chainID)),
		hx.Target("#" + containerID),
		hx.Swap("outerHTML"),
		hx.Trigger("change"),
		hx.Include("#" + containerID),
	}
	if checked {
		attrs = append(attrs, g.Attr("checked", "checked"))
	}
	return attrs
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

func priceCurrencySelect(selected string) g.Node {
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
		ID("price-currency"),
		Class("w-24 p-2 border rounded-md shrink-0"),
		g.Group(opts),
	)
}

func textInput(f field.ChainField, id string, spec field.InputSpec) g.Node {
	attrs := []g.Node{
		Type("text"),
		Name(f.FieldName),
		ID("input-" + id),
		Class("w-full p-2 border rounded-md"),
	}
	return Input(appendValidationAttrs(attrs, f, spec)...)
}

func textareaInput(f field.ChainField, id string, spec field.InputSpec) g.Node {
	attrs := []g.Node{
		Name(f.FieldName),
		ID("input-" + id),
		Class("w-full p-2 border rounded-md"),
		g.Attr("rows", "6"),
	}
	return Textarea(appendTextareaValidationAttrs(attrs, f, spec)...)
}

func fieldLabel(f field.ChainField, forID string) g.Node {
	return Label(For(forID), Class("field-label"), g.Text(fieldLabelText(f)))
}

func specSelect(categoryID int, chain field.ChainGroup, f field.ChainField, options []string) g.Node {
	selectID := "select-" + fieldContainerID(chain.ChainID, f.FieldName)
	nextField := nextChainFieldName(chain, f.FieldName)

	selectAttrs := []g.Node{
		ID(selectID),
		Name(f.FieldName),
		Class("w-full p-2 border rounded-md"),
	}
	if f.IsRequired {
		selectAttrs = append(selectAttrs, g.Attr("required", "required"))
	}

	optNodes := []g.Node{Option(Value(""), g.Text("Select…"))}
	for _, o := range options {
		optNodes = append(optNodes, Option(Value(o), g.Text(o)))
	}

	if nextField != "" {
		selectAttrs = append(selectAttrs,
			hx.Get(nextFieldURL(categoryID, chain.ChainID, f.FieldName, nextField)),
			hx.Target("#"+chainNextID(chain.ChainID, f.FieldName)),
			hx.Swap("outerHTML"),
			hx.Trigger("change"),
			hx.Include("#"+chainContainerID(chain.ChainID)),
		)
	}

	return Select(append(selectAttrs, optNodes...)...)
}

func specMultiCheckboxGrid(categoryID int, chain field.ChainGroup, f field.ChainField, options []string) g.Node {
	gridID := "select-" + fieldContainerID(chain.ChainID, f.FieldName)
	nextField := nextChainFieldName(chain, f.FieldName)

	gridAttrs := []g.Node{
		ID(gridID),
		Class("grid grid-cols-4 sm:grid-cols-6 gap-2 border rounded-md p-3"),
		g.Attr("role", "group"),
	}
	if nextField != "" {
		gridAttrs = append(gridAttrs,
			hx.Get(nextFieldURL(categoryID, chain.ChainID, f.FieldName, nextField)),
			hx.Target("#"+chainNextID(chain.ChainID, f.FieldName)),
			hx.Swap("outerHTML"),
			hx.Trigger("change"),
			hx.Include("#"+chainContainerID(chain.ChainID)),
		)
	}

	for _, o := range options {
		gridAttrs = append(gridAttrs, checkboxGridItem(f.FieldName, o, false))
	}
	return Div(gridAttrs...)
}

// checkboxGridItem renders one labeled checkbox for a multi-value field grid.
func checkboxGridItem(name, value string, checked bool) g.Node {
	inputAttrs := []g.Node{
		Type("checkbox"),
		Name(name),
		Value(value),
	}
	if checked {
		inputAttrs = append(inputAttrs, g.Attr("checked", "checked"))
	}
	return Label(
		Class("checkbox-option flex items-center gap-2"),
		Input(inputAttrs...),
		Span(Class("checkbox-label"), g.Text(value)),
	)
}

func nextFieldURL(categoryID, chainID int, afterField, fieldName string) string {
	v := url.Values{}
	v.Set("category_id", strconv.Itoa(categoryID))
	v.Set("chain_id", strconv.Itoa(chainID))
	v.Set("after", afterField)
	v.Set("field", fieldName)
	return "/api/ad/new/next-field?" + v.Encode()
}

func priceFieldURL(categoryID, chainID int) string {
	v := url.Values{}
	v.Set("category_id", strconv.Itoa(categoryID))
	v.Set("chain_id", strconv.Itoa(chainID))
	return "/api/ad/new/price-field?" + v.Encode()
}

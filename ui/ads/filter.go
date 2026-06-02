package ads

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rocky-ads/site/field"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
	hx "maragu.dev/gomponents-htmx"
)

type FilterView struct {
	CategoryID int
	Category   field.CategoryChains
	OptionsMap map[int]map[string][]string
	Filters    field.AdFilters
}

// FilterContent renders filter fields for the site search widget (parent form handles submit).
func FilterContent(view FilterView) g.Node {
	fields := filterFieldNodes(view)
	if len(fields) == 0 {
		return g.Group{}
	}
	return Div(Class("col-span-2 grid grid-cols-2 gap-4"), g.Group(fields))
}

func AdFiltersPartial(view FilterView) g.Node {
	fields := filterFieldNodes(view)
	if len(fields) == 0 {
		return Div(ID("ad-filters"), Class("ad-filters ad-filters-empty"))
	}

	children := []g.Node{
		ID("ad-filters"),
		Class("ad-filters"),
		hx.Get("/ads/list"),
		hx.Target("#ad-list"),
		hx.Trigger("change delay:300ms"),
		hx.Include("#home_category_id"),
		H2(Class("ad-filters-title"), g.Text("Filters")),
		Div(Class("ad-filters-fields"), g.Group(fields)),
	}
	if view.Filters.HasFilters() {
		children = append(children, clearFiltersLink(view.CategoryID))
	}
	return Form(children...)
}

func clearFiltersLink(categoryID int) g.Node {
	return A(
		Class("ad-filters-clear"),
		Href("/home/content?category_id="+strconv.Itoa(categoryID)),
		g.Attr("hx-get", "/home/content?category_id="+strconv.Itoa(categoryID)),
		g.Attr("hx-target", "#home-content"),
		g.Attr("hx-swap", "outerHTML"),
		g.Text("Clear filters"),
	)
}

func filterFieldNodes(view FilterView) []g.Node {
	var plainFields []g.Node
	var chainColumns []g.Node

	for _, chain := range view.Category.Chains {
		if field.IsSpecChain(chain) {
			if section := filterSpecChainFields(view, chain); len(section) > 0 {
				chainColumns = append(chainColumns, Div(
					ID(filterChainContainerID(chain.ChainID)),
					Class("ad-filter-chain"),
					g.Group(section),
				))
			}
			continue
		}
		for _, f := range chain.Fields {
			if field.IsFilterExcluded(f.FieldName) {
				continue
			}
			if row := filterPlainFieldRow(view.Filters, f); row != nil {
				plainFields = append(plainFields, row)
			}
		}
	}

	var nodes []g.Node
	if len(plainFields) > 0 {
		nodes = append(nodes, Div(Class("ad-filter-plain-column"), g.Group(plainFields)))
	}
	nodes = append(nodes, chainColumns...)
	return nodes
}

func filterSpecChainFields(view FilterView, chain field.ChainGroup) []g.Node {
	chainOpts := view.OptionsMap[chain.ChainID]
	if len(chainOpts) == 0 {
		return nil
	}

	var nodes []g.Node
	var afterField string
	for _, f := range chain.Fields {
		if field.IsFilterExcluded(f.FieldName) {
			continue
		}
		opts, visible := chainOpts[f.FieldName]
		if !visible {
			break
		}
		if opts == nil {
			if afterField != "" {
				nodes = append(nodes, filterNoMatchPartial(chain, afterField, f.FieldName))
			}
			return nodes
		}
		if len(opts) == 0 {
			return nodes
		}

		nodes = append(nodes, filterSpecFieldRow(view.CategoryID, chain, f, opts, view.Filters.Values(f.FieldName)))
		afterField = f.FieldName

		selected := view.Filters.Values(f.FieldName)
		if len(selected) == 0 {
			nodes = append(nodes, filterChainNextPlaceholder(chain.ChainID, f.FieldName))
			return nodes
		}
	}

	if afterField != "" && filterNextChainFieldName(chain, afterField) != "" {
		nodes = append(nodes, filterChainNextPlaceholder(chain.ChainID, afterField))
	}
	return nodes
}

// FilterNextFieldPartial swaps into the filter chain placeholder after a selection.
func FilterNextFieldPartial(categoryID int, chain field.ChainGroup, afterField, fieldName string, options []string, noMatch bool) g.Node {
	if fieldName == "" {
		return filterChainNextPlaceholder(chain.ChainID, afterField)
	}

	if noMatch || len(options) == 0 {
		return filterNoMatchPartial(chain, afterField, fieldName)
	}

	f, ok := filterFindChainField(chain, fieldName)
	if !ok {
		return filterChainNextPlaceholder(chain.ChainID, afterField)
	}

	nextAfter := filterNextChainFieldName(chain, fieldName)
	children := []g.Node{filterSpecFieldRow(categoryID, chain, f, options, nil)}
	if nextAfter != "" {
		children = append(children, filterChainNextPlaceholder(chain.ChainID, f.FieldName))
	}

	return Div(
		ID(filterChainNextID(chain.ChainID, afterField)),
		Class("chain-next"),
		g.Group(children),
	)
}

func filterNoMatchPartial(chain field.ChainGroup, afterField, fieldName string) g.Node {
	child, okChild := filterFindChainField(chain, fieldName)
	parent, okParent := filterFindChainField(chain, afterField)
	childLabel := fieldName
	parentLabel := afterField
	if okChild {
		childLabel = filterPluralDisplayName(child.DisplayName)
	}
	if okParent {
		parentLabel = parent.DisplayName
	}
	msg := fmt.Sprintf("No %s in current ads match the %s selection.", childLabel, parentLabel)

	return Div(
		ID(filterChainNextID(chain.ChainID, afterField)),
		Class("chain-next"),
		P(Class("field-no-match"), g.Text(msg)),
	)
}

func filterChainNextPlaceholder(chainID int, afterField string) g.Node {
	return Div(ID(filterChainNextID(chainID, afterField)), Class("chain-next"))
}

func filterSpecFieldRow(categoryID int, chain field.ChainGroup, f field.ChainField, options, selected []string) g.Node {
	return Div(
		ID(filterFieldContainerID(chain.ChainID, f.FieldName)),
		Class("mt-3"),
		Label(For("filter-"+f.FieldName), Class("field-label"), g.Text(f.DisplayName)),
		filterSpecFieldInput(categoryID, chain, f, options, selected),
	)
}

func filterSpecFieldInput(categoryID int, chain field.ChainGroup, f field.ChainField, options, selected []string) g.Node {
	if field.IsMultiInput(f.InputType) {
		return filterMultiCheckboxGrid(categoryID, chain, f, options, selected)
	}

	nextField := filterNextChainFieldName(chain, f.FieldName)
	selectID := "filter-" + f.FieldName

	selectedSet := selectedSet(selected)

	selectAttrs := []g.Node{
		ID(selectID),
		Name(f.FieldName),
		Class("w-full p-2 border rounded-md"),
	}

	if nextField != "" {
		selectAttrs = append(selectAttrs,
			hx.Get(filterNextFieldURL(categoryID, chain.ChainID, f.FieldName, nextField)),
			hx.Target("#"+filterChainNextID(chain.ChainID, f.FieldName)),
			hx.Swap("outerHTML"),
			hx.Trigger("change"),
			hx.Include("#"+filterChainContainerID(chain.ChainID)),
		)
	}

	optNodes := []g.Node{Option(Value(""), g.Text("Any"))}
	for _, o := range options {
		attrs := []g.Node{Value(o), g.Text(o)}
		if _, ok := selectedSet[o]; ok {
			attrs = append(attrs, g.Attr("selected", "selected"))
		}
		optNodes = append(optNodes, Option(attrs...))
	}

	return Select(append(selectAttrs, optNodes...)...)
}

func filterMultiCheckboxGrid(categoryID int, chain field.ChainGroup, f field.ChainField, options, selected []string) g.Node {
	gridID := "filter-" + f.FieldName
	nextField := filterNextChainFieldName(chain, f.FieldName)
	set := selectedSet(selected)

	gridAttrs := []g.Node{
		ID(gridID),
		Class("checkbox-grid"),
		g.Attr("role", "group"),
	}
	if nextField != "" {
		gridAttrs = append(gridAttrs,
			hx.Get(filterNextFieldURL(categoryID, chain.ChainID, f.FieldName, nextField)),
			hx.Target("#"+filterChainNextID(chain.ChainID, f.FieldName)),
			hx.Swap("outerHTML"),
			hx.Trigger("change"),
			hx.Include("#"+filterChainContainerID(chain.ChainID)),
		)
	}

	for _, o := range options {
		_, checked := set[o]
		gridAttrs = append(gridAttrs, checkboxGridItem(f.FieldName, o, checked))
	}
	return Div(gridAttrs...)
}

func selectedSet(selected []string) map[string]struct{} {
	set := make(map[string]struct{}, len(selected))
	for _, v := range selected {
		set[v] = struct{}{}
	}
	return set
}

func filterPlainFieldRow(filters field.AdFilters, f field.ChainField) g.Node {
	spec := field.ParseInputType(f.InputType)
	switch f.FieldName {
	case "price":
		return filterPriceRow(filters)
	case "location":
		return filterLocationRow(filters)
	}
	switch spec.Type {
	case field.InputNumber:
		return filterNumberRow(f, filters.Values(f.FieldName))
	default:
		return filterTextRow(f, filters.Values(f.FieldName))
	}
}

func filterPriceRow(filters field.AdFilters) g.Node {
	minVal := priceAmount(filters.PriceMin)
	maxVal := priceAmount(filters.PriceMax)
	return Div(
		Class("mt-3"),
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
				g.If(minVal != "", Value(minVal)),
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
				g.If(maxVal != "", Value(maxVal)),
			),
		),
	)
}

func filterLocationRow(filters field.AdFilters) g.Node {
	attrs := []g.Node{
		Type("text"),
		Name("location"),
		ID("filter-location"),
		Class("w-full p-2 border rounded-md"),
		g.Attr("placeholder", "City or state"),
	}
	if filters.Location != "" {
		attrs = append(attrs, Value(filters.Location))
	}
	return Div(
		Class("mt-3"),
		Label(For("filter-location"), Class("field-label"), g.Text("Location")),
		Input(attrs...),
	)
}

func filterNumberRow(f field.ChainField, selected []string) g.Node {
	id := "filter-" + f.FieldName
	val := ""
	if len(selected) > 0 {
		val = selected[0]
	}
	attrs := []g.Node{
		Type("number"),
		Name(f.FieldName),
		ID(id),
		Class("w-full p-2 border rounded-md"),
		g.Attr("min", "0"),
		g.Attr("inputmode", "numeric"),
	}
	if val != "" {
		attrs = append(attrs, Value(val))
	}
	return Div(
		Class("mt-3"),
		Label(For(id), Class("field-label"), g.Text(f.DisplayName)),
		Input(attrs...),
	)
}

func filterTextRow(f field.ChainField, selected []string) g.Node {
	id := "filter-" + f.FieldName
	val := ""
	if len(selected) > 0 {
		val = selected[0]
	}
	attrs := []g.Node{
		Type("text"),
		Name(f.FieldName),
		ID(id),
		Class("w-full p-2 border rounded-md"),
	}
	if val != "" {
		attrs = append(attrs, Value(val))
	}
	return Div(
		Class("mt-3"),
		Label(For(id), Class("field-label"), g.Text(f.DisplayName)),
		Input(attrs...),
	)
}

func filterNextFieldURL(categoryID, chainID int, afterField, fieldName string) string {
	v := url.Values{}
	v.Set("category_id", strconv.Itoa(categoryID))
	v.Set("chain_id", strconv.Itoa(chainID))
	v.Set("after", afterField)
	v.Set("field", fieldName)
	return "/api/filter/next-field?" + v.Encode()
}

func filterFindChainField(chain field.ChainGroup, fieldName string) (field.ChainField, bool) {
	for _, f := range chain.Fields {
		if f.FieldName == fieldName {
			return f, true
		}
	}
	return field.ChainField{}, false
}

func filterNextChainFieldName(chain field.ChainGroup, current string) string {
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

func filterPluralDisplayName(displayName string) string {
	if strings.HasSuffix(displayName, "s") {
		return displayName
	}
	if strings.HasSuffix(displayName, "y") {
		return displayName[:len(displayName)-1] + "ies"
	}
	return displayName + "s"
}

func priceAmount(amount *int) string {
	if amount == nil {
		return ""
	}
	return strconv.Itoa(*amount)
}

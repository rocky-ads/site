package ui

import (
	uiads "github.com/rocky-ads/site/ui/ads"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func HomePage(userID, view int, categoryName, categoryImage, q string, filtersExpanded bool, filters uiads.SearchFilters, results []g.Node) []g.Node {
	return []g.Node{SearchContainer(userID, view, categoryName, categoryImage, q, filtersExpanded, filters, results)}
}

func SearchContainer(userID, view int, categoryName, categoryImage, q string, filtersExpanded bool, filters uiads.SearchFilters, results []g.Node) g.Node {
	return g.Group(append([]g.Node{
		Div(
			ID("search-container"),
			Div(
				categoryButton(categoryName, categoryImage),
				SearchWidget(userID, view, q, filtersExpanded, filters, results),
			),
		),
	}, RemoveModal("category")...))
}

func SearchResults(view int, results []g.Node) g.Node {
	return searchResults(view, results, false)
}

// SearchResultsResponse is the HTMX body for /api/search/.
// Page 1: #search-results with sorry message or first page of ads.
// Page 2+: ad nodes only (replaces the scroll sentinel).
func SearchResultsResponse(view, page int, results []g.Node) g.Node {
	if page > 1 {
		return g.Group(results)
	}
	return SearchResults(view, results)
}

// SearchResultsOOB swaps #search-results out-of-band (show/hide filters).
func SearchResultsOOB(view int, results []g.Node) g.Node {
	return searchResults(view, results, true)
}

// FilterPanel is the HTMX fragment inserted into #filter-panel when expanded.
func FilterPanel(filters uiads.SearchFilters) g.Node {
	return Div(
		Class("border rounded-lg p-4 mt-4"),
		uiads.SearchFiltersPanel(filters),
	)
}

func categoryButton(categoryName, categoryImage string) g.Node {
	imagePath := "/images/category/" + categoryImage

	return Div(
		Div(
			Class("flex items-center gap-5 mb-4"),
			Label(
				Class("font-bold"),
				g.Text("Category"),
			),
			Button(
				Type("button"),
				Class("py-2 px-5 flex items-center gap-2 rounded-full border-2 border-blue-500 bg-blue-100 hover:bg-blue-200 dark:bg-blue-900 dark:hover:bg-blue-800 dark:border-blue-400"),
				hx.Get("/api/category-select?return=%2F"),
				hx.Target("body"),
				hx.Swap("beforeend"),
				Img(
					Src(imagePath),
					Alt("Category icon"),
					Class("w-6 h-6 dark:invert dark:opacity-80"),
				),
				Span(g.Text(categoryName)),
			),
		),
	)
}

func searchBox(q string) g.Node {
	return Input(
		Class("w-full p-2 border rounded"),
		Type("search"),
		ID("searchBox"),
		Name("q"),
		Value(q),
		hx.Trigger("search"),
		Placeholder("What are you looking for?"),
	)
}

func FilterToggle(expanded bool) g.Node {
	return filterToggle(expanded, false)
}

// FilterToggleOOB swaps #filter-toggle when the panel is shown or hidden.
func FilterToggleOOB(expanded bool) g.Node {
	return filterToggle(expanded, true)
}

func filterToggle(expanded bool, oob bool) g.Node {
	label := "Expand filters"
	icon := "/images/expand.svg"
	var actionAttrs []g.Node
	if expanded {
		label = "Collapse filters"
		icon = "/images/collapse.svg"
		actionAttrs = []g.Node{
			hx.Get("/api/hide-filters"),
			hx.Target("#filter-panel"),
			hx.Swap("innerHTML"),
			hx.Include("#search-widget"),
		}
	} else {
		actionAttrs = []g.Node{
			hx.Get("/api/show-filters"),
			hx.Target("#filter-panel"),
			hx.Swap("innerHTML"),
		}
	}
	attrs := []g.Node{
		Type("button"),
		ID("filter-toggle"),
		Class("p-2 border border-zinc-300 dark:border-zinc-600 rounded-md hover:bg-zinc-100 dark:hover:bg-zinc-800"),
		g.Attr("aria-label", label),
		g.Attr("title", label),
	}
	if oob {
		attrs = append(attrs, hx.SwapOOB("outerHTML"))
	}
	attrs = append(attrs, actionAttrs...)
	attrs = append(attrs, Img(
		Class("w-6 h-6 dark:invert dark:opacity-80"),
		Src(icon),
		Alt(label),
	))
	return Button(attrs...)
}

func SearchView(userID, view int, results []g.Node) g.Node {
	return Div(
		Class("flex flex-col gap-4"),
		ID("search-view"),
		viewRow(userID, view),
		searchResults(view, results, false),
	)
}

func viewRow(userID, view int) g.Node {
	return Div(
		Class("flex justify-between items-center gap-2 my-4"),
		Div(
			Class("flex gap-2"),
			newAdButton(userID),
		),
		viewToggles(view),
	)
}

func newAdButton(userID int) g.Node {
	return standardButton(buttonProps{
		Href:     "/auth/ad/new",
		Text:     "New Ad",
		Disabled: userID == 0,
	})
}

func SearchWidget(userID, view int, q string, filtersExpanded bool, filters uiads.SearchFilters, results []g.Node) g.Node {
	var panel g.Node
	if filtersExpanded {
		panel = FilterPanel(filters)
	}
	return Form(
		Class("flex flex-col gap-4"),
		ID("search-widget"),
		hx.Get("/api/search/"),
		hx.Target("#search-results"),
		hx.Swap("outerHTML"),
		hx.Include("#search-widget"),
		hx.Trigger("search, keydown[key=='Tab'] from:#searchBox, change from:#filter-panel input delay:300ms, change from:#filter-panel select delay:300ms, keydown[key=='Enter'] from:#filter-location"),
		Div(
			ID("search-bar"),
			Class("flex gap-2 items-center"),
			searchBox(q),
			FilterToggle(filtersExpanded),
		),
		Div(ID("filter-panel"), panel),
		SearchView(userID, view, results),
	)
}

func searchResults(view int, results []g.Node, oob bool) g.Node {
	var class string

	switch view {
	case ViewGrid:
		class = "grid grid-cols-2 md:grid-cols-3 gap-3"
	case ViewList:
		// Empty class - items naturally stack as a column
	}

	var content g.Node
	if len(results) == 0 {
		content = searchResultsEmpty()
	} else {
		content = g.Group(results)
	}

	attrs := []g.Node{
		ID("search-results"),
		Class(class),
		content,
	}
	if oob {
		attrs = append([]g.Node{hx.SwapOOB("outerHTML")}, attrs...)
	}
	return Div(attrs...)
}

func searchResultsEmpty() g.Node {
	return P(
		Class("col-span-full py-8 text-center text-zinc-500 dark:text-zinc-400"),
		g.Text("Sorry, no ads found matching that criteria."),
	)
}

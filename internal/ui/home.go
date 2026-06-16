package ui

import (
	"net/url"
	"strconv"

	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/local"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func HomePage(userID, view int, q string, filtersExpanded bool, category CategoryOption, filterFacets []facet.Def, filters uiads.SearchFilters, results []g.Node) []g.Node {
	return []g.Node{SearchContainer(userID, view, q, filtersExpanded, category, filterFacets, filters, results)}
}

func SearchContainer(userID, view int, q string, filtersExpanded bool, category CategoryOption, filterFacets []facet.Def, filters uiads.SearchFilters, results []g.Node) g.Node {
	return g.Group(append([]g.Node{
		Div(
			ID("search-container"),
			Div(
				categoryButton(category, "/"),
				SearchWidget(userID, view, q, filtersExpanded, category, filterFacets, filters, results),
			),
		),
	}, append(RemoveModal("category"), RemoveModal("search-location")...)...))
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
func FilterPanel(
	q string,
	filterFacets []facet.Def,
	filters uiads.SearchFilters,
) g.Node {
	return Div(
		Class("border rounded-lg p-4 mt-4"),
		Div(
			Class("flex gap-2 items-center mb-4"),
			Div(Class("flex-1 min-w-0"), searchBox(q)),
			FilterToggle(true),
		),
		uiads.SearchFiltersPanel(filterFacets, filters),
	)
}

func searchBarRow(q string, expanded bool) g.Node {
	if expanded {
		return Div(ID("search-bar"), Class("hidden"))
	}
	return Div(
		ID("search-bar"),
		Class("flex gap-2 items-center"),
		Div(Class("flex-1 min-w-0"), searchBox(q)),
		FilterToggle(false),
	)
}

// SearchBarOOB swaps #search-bar when the filter panel is shown or hidden.
func SearchBarOOB(q string, expanded bool) g.Node {
	if expanded {
		return Div(
			ID("search-bar"),
			Class("hidden"),
			hx.SwapOOB("outerHTML"),
		)
	}
	return Div(
		ID("search-bar"),
		Class("flex gap-2 items-center"),
		hx.SwapOOB("outerHTML"),
		Div(Class("flex-1 min-w-0"), searchBox(q)),
		FilterToggle(false),
	)
}

func categoryButton(category CategoryOption, returnParam string) g.Node {
	imagePath := "/images/category/" + category.ImageFile

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
				hx.Get("/api/category-select?return="+url.QueryEscape(returnParam)),
				hx.Target("body"),
				hx.Swap("beforeend"),
				Img(
					Src(imagePath),
					Alt("Category icon"),
					Class("w-6 h-6 dark:invert dark:opacity-80"),
				),
				Span(g.Text(category.Name)),
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
	return filterToggle(expanded)
}

func filterToggle(expanded bool) g.Node {
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
		Disabled: !local.IsLoggedIn(userID),
	})
}

func SearchWidget(userID, view int, q string, filtersExpanded bool, category CategoryOption, filterFacets []facet.Def, filters uiads.SearchFilters, results []g.Node) g.Node {
	var panel g.Node
	if filtersExpanded {
		panel = FilterPanel(q, filterFacets, filters)
	}
	locationBar := uiads.SearchLocationBar(filters)

	attrs := []g.Node{
		Class("flex flex-col gap-4"),
		ID("search-widget"),
		g.Attr("onsubmit", "event.preventDefault(); return false;"),
		hx.Get("/api/search/"),
		hx.Target("#search-results"),
		hx.Swap("outerHTML"),
		hx.Include("#search-widget"),
		hx.Trigger("search, keydown[key=='Tab'] from:#searchBox, change from:(#filter-panel input) delay:300ms, change from:(#filter-panel select) delay:300ms"),
	}

	if filtersExpanded {
		attrs = append(attrs,
			searchBarRow(q, true),
			Div(ID("filter-panel"), panel),
			locationBar,
		)
	} else {
		attrs = append(attrs,
			Div(
				Class("flex flex-col gap-1"),
				searchBarRow(q, false),
				locationBar,
			),
			Div(ID("filter-panel")),
		)
	}
	attrs = append(attrs, SearchView(userID, view, results))
	return Form(attrs...)
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
	return SearchResultsEmptyMessage()
}

// SearchResultsEmptyMessage is the standard no-results copy for search grids.
func SearchResultsEmptyMessage() g.Node {
	return P(
		Class("col-span-full py-8 text-center text-zinc-500 dark:text-zinc-400"),
		g.Text("Sorry, no ads found matching that criteria."),
	)
}

// OutsideAreaHeading separates in-area from out-of-area search results.
func OutsideAreaHeading() g.Node {
	return H2(
		Class("col-span-full text-lg font-semibold mt-4 mb-2"),
		g.Text("Outside of area"),
	)
}

// NoInAreaMatchesMessage is shown when geo search has no matches in the within area.
func NoInAreaMatchesMessage(within int, unit, location string) g.Node {
	suffix := " mi"
	if unit == "km" {
		suffix = " km"
	}
	msg := "No matching ads were found within " +
		strconv.Itoa(within) + suffix + " of " + location
	return P(
		Class("col-span-full py-4 text-center text-zinc-500 dark:text-zinc-400"),
		g.Text(msg),
	)
}

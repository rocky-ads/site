package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func HomePage(userID, view int, categoryName, categoryImage string, results []g.Node) []g.Node {
	return []g.Node{SearchContainer(userID, view, categoryName, categoryImage, results)}
}

func SearchContainer(userID, view int, categoryName, categoryImage string, results []g.Node) g.Node {
	return g.Group(append([]g.Node{
		Div(
			ID("search-container"),
			categorySearch(userID, view, categoryName, categoryImage, results),
		),
	}, RemoveModal("category")...))
}

func categorySearch(userID, view int, categoryName, categoryImage string, results []g.Node) g.Node {
	return Div(
		categoryButton(categoryName, categoryImage),
		SearchWidget(userID, view, "", results, []g.Node{}),
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

func filtersButton() g.Node {
	return standardButton(buttonProps{
		Type: "button",
		Text: "Filters",
		Attrs: []g.Node{
			hx.Get("/api/show-filters"),
			hx.Target("#search-widget"),
			hx.Swap("outerHTML"),
		},
	})
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

func clearFilters() g.Node {
	return standardButton(buttonProps{
		Type: "button",
		Text: "Clear",
		Attrs: []g.Node{
			hx.Get("/api/show-filters"),
			hx.Target("#search-widget"),
			hx.Swap("outerHTML"),
			hx.Params("none"),
		},
	})
}

func applyFilters() g.Node {
	return standardButton(buttonProps{
		Type: "submit",
		Text: "Apply",
	})
}

func filterActions() g.Node {
	return Div(
		Class("flex justify-end gap-2 mt-4"),
		clearFilters(),
		applyFilters(),
	)
}

func filterControls(filters []g.Node) g.Node {
	return Div(
		Class("grid grid-cols-2 gap-4 mt-4"),
		g.Group(filters),
	)
}

func searchFilters(q string, filters []g.Node) g.Node {
	return Div(
		Class("border rounded-lg p-4"),
		searchBox(q),
		filterControls(filters),
		filterActions(),
	)
}

func searchSimple(q string) g.Node {
	return Div(
		Class("flex gap-2 items-center"),
		searchBox(q),
		filtersButton(),
	)
}

func newAdButton(userID int) g.Node {
	return standardButton(buttonProps{
		Href:     "/auth/ad/new",
		Text:     "New Ad",
		Disabled: userID == 0,
	})
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

func SearchView(userID, view int, results []g.Node) g.Node {
	return Div(
		Class("flex flex-col gap-4"),
		ID("search-view"),
		viewRow(userID, view),
		searchResults(view, results),
	)
}

func SearchWidget(userID, view int, q string, results, filters []g.Node) g.Node {
	return Form(
		Class("flex flex-col gap-4"),
		ID("search-widget"),
		hx.Get("/api/search"),
		hx.Target("#search-results"),
		hx.Swap("outerHTML"),
		hx.Include("form"),
		g.If(len(filters) > 0, searchFilters(q, filters)),
		g.If(len(filters) == 0, searchSimple(q)),
		SearchView(userID, view, results),
	)
}

func searchResults(view int, results []g.Node) g.Node {
	var class string

	switch view {
	case ViewGrid:
		class = "grid grid-cols-2 md:grid-cols-3 gap-3"
	case ViewList:
		// Empty class - items naturally stack as a column
	}

	return Div(
		ID("search-results"),
		Class(class),
		g.Group(results),
	)
}

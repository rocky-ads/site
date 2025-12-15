package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func HomePage(categoryName, categoryImage string) []g.Node {
	return []g.Node{SearchContainer(categoryName, categoryImage)}
}

func SearchContainer(categoryName, categoryImage string) g.Node {
	return Div(
		ID("search-container"),
		categorySearch(categoryName, categoryImage),
	)
}

func categorySearch(categoryName, categoryImage string) g.Node {
	return Div(
		categoryButton(categoryName, categoryImage),
		searchWidget(),
	)
}

func categoryButton(categoryName, categoryImage string) g.Node {
	imagePath := "/images/category/" + categoryImage

	return Div(
		Div(
			Class("flex items-center gap-5"),
			Label(
				Class("font-bold"),
				g.Text("Category"),
			),
			Button(
				Type("button"),
				Class("py-2 px-5 flex items-center gap-2 rounded-full border-2 border-blue-500 bg-blue-100 hover:bg-blue-200 dark:bg-blue-900 dark:hover:bg-blue-800 dark:border-blue-400"),
				hx.Get("/api/modal/category-select"),
				hx.Target("body"),
				hx.Swap("beforeend"),
				Img(
					Src(imagePath),
					Alt("Category icon"),
					Class("w-6 h-6 dark:invert"),
				),
				Span(g.Text(categoryName)),
			),
		),
	)
}

func searchWidget() g.Node {
	return Form(
		Class("flex flex-col gap-4"),
		Method("GET"),
		hx.Get("/api/search"),
		hx.Target("#search-results"),
		hx.Swap("outerHTML"),
		hx.Include("form"),
		ID("search-widget"),
		Input(
			ID("search-input"),
			Name("q"),
			Type("text"),
			Placeholder("What are you looking for?"),
		),
		searchResults(),
	)
}

func searchResults() g.Node {
	return Div(
		ID("search-results"),
		searchResultsList(),
	)
}

func searchResultsList() g.Node {
	return Div(
		ID("search-results-list"),
	)
}

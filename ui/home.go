package ui

import (
	"github.com/rocky-ads/site/models"
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
		searchWidget("", make(models.FieldValues), false),
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

func filtersButton() g.Node {
	return StandardButton(ButtonProps{
		Type: "button",
		Text: "Filters",
		Attrs: []g.Node{
			hx.Get("/search-widget?show-filters=true"),
			hx.Target("#searchWidget"),
			hx.Swap("outerHTML"),
			hx.Include("form"),
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

func searchFilters(q string, fv models.FieldValues) g.Node {
	return Div(
		Class("border rounded-lg p-4"),
		searchBox(q),
		//filterControls(fv),
		//filterActions(fv),
	)
}

func searchSimple(q string) g.Node {
	return Div(
		Class("flex gap-2 items-center"),
		searchBox(q),
		filtersButton(),
	)
}

func searchWidget(q string, fv models.FieldValues, showFilters bool) g.Node {
	return Form(
		Class("flex flex-col gap-4"),
		ID("search-widget"),
		hx.Get("/api/search"),
		hx.Target("#search-results"),
		hx.Swap("outerHTML"),
		hx.Include("form"),
		g.If(showFilters, searchFilters(q, fv)),
		g.If(!showFilters, searchSimple(q)),
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

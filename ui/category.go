package ui

import (
	"net/url"
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func CategoryItem(currentCategoryID, categoryID int, name, imageFile, returnParam string) g.Node {
	itemClass := "flex items-center gap-3 p-3 hover:bg-zinc-50 dark:hover:bg-zinc-800 cursor-pointer rounded-lg transition-colors text-zinc-900 dark:text-zinc-200 "
	if categoryID == currentCategoryID {
		itemClass += "bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-700"
	}

	returnURL := "/api/category/" + strconv.Itoa(categoryID) + "/switch?return=" + url.QueryEscape(returnParam)

	return Div(
		Class(itemClass),
		hx.Get(returnURL),
		hx.Swap("none"),
		Div(
			Class("p-2 bg-zinc-200 dark:bg-zinc-700 rounded-full flex items-center justify-center"),
			Img(
				Src("/images/category/"+imageFile),
				Alt("Category icon"),
				Class("w-6 h-6 dark:invert dark:opacity-80"),
			),
		),
		Span(Class("flex-1"), g.Text(name)),
	)
}

func CategorySelectModal(categoryItems []g.Node) g.Node {
	return g.Group([]g.Node{
		modalBackdrop("category"),
		Div(
			ID("category-modal"),
			Class("fixed inset-0 flex items-center justify-center z-50 p-8 pointer-events-none"),
			Div(
				Class("bg-white dark:bg-zinc-800 rounded-lg w-full shadow-2xl border-2 border-zinc-300 dark:border-zinc-600 flex flex-col pointer-events-auto"),
				Style("max-width: 400px; max-height: 80vh"),
				Div(
					Class("flex items-center justify-between p-6 border-b border-zinc-200 dark:border-zinc-700 flex-shrink-0"),
					H3(Class("text-xl font-bold text-zinc-900 dark:text-zinc-200"), g.Text("Select Category")),
					modalClose("category"),
				),
				Div(
					Class("flex-1 overflow-y-auto p-6 pt-4"),
					Div(
						Class("space-y-2"),
						g.Group(categoryItems),
					),
				),
			),
		),
	})
}

package ui

import (
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func CategoryItem(currentCategoryID, categoryID int, name, imageFile string) g.Node {
	itemClass := "flex items-center gap-3 p-3 hover:bg-gray-50 cursor-pointer rounded-lg transition-colors "
	if categoryID == currentCategoryID {
		itemClass += "bg-blue-50 border border-blue-200"
	}

	return Div(
		Class(itemClass),
		hx.Get("/api/category/"+strconv.Itoa(categoryID)+"/switch"),
		hx.Target("#search-container"),
		hx.Swap("outerHTML"),
		hideModalOnClick,
		Div(
			Class("p-2 bg-gray-200 rounded-full flex items-center justify-center"),
			Img(
				Src("/images/category/"+imageFile),
				Alt("Category icon"),
				Class("w-6 h-6"),
			),
		),
		Span(Class("text-gray-700 flex-1"), g.Text(name)),
	)
}

func CategorySelectModal(categoryItems []g.Node) g.Node {
	return g.Group([]g.Node{
		modalBackdrop(),
		Div(
			ID("modal"),
			hx.SwapOOB("true"),
			Class("modal fixed inset-0 flex items-center justify-center z-50 p-8 pointer-events-none"),
			Div(
				Class("bg-white rounded-lg w-full shadow-2xl border-2 border-gray-300 flex flex-col pointer-events-auto"),
				Style("max-width: 400px; max-height: 80vh"),
				Div(
					Class("flex items-center justify-between p-6 border-b border-gray-200 flex-shrink-0"),
					H3(Class("text-xl font-bold text-gray-900"), g.Text("Select Category")),
					Button(
						Type("button"),
						Class("bg-white border-2 border-gray-800 rounded-full w-8 h-8 flex items-center justify-center shadow-lg hover:bg-gray-100 focus:outline-none cursor-pointer"),
						hideModalOnClick,
						Img(
							Src("/images/close.svg"),
							Alt("Close"),
							Class("w-4 h-4"),
						),
					),
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

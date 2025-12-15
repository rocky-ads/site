package ui

import (
	"strconv"

	"github.com/rocky-ads/site/models"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func createCategoryItems(categoryID int, categories []models.Category) []g.Node {
	var cats []g.Node

	for _, cat := range categories {
		itemClass := "flex items-center gap-3 p-3 hover:bg-gray-50 cursor-pointer rounded-lg transition-colors "
		if cat.ID == categoryID {
			itemClass += "bg-blue-50 border border-blue-200"
		}

		item := Div(
			Class(itemClass),
			hx.Get("/api/category/"+strconv.Itoa(cat.ID)+"/switch"),
			hx.Target("#search-container"),
			hx.Swap("outerHTML"),
			Div(
				Class("p-2 bg-gray-200 rounded-full flex items-center justify-center"),
				Img(
					Src("/images/category/"+cat.ImageFile),
					Alt("Category icon"),
					Class("w-6 h-6"),
				),
			),
			Span(Class("text-gray-700 flex-1"), g.Text(cat.Name)),
		)
		cats = append(cats, item)
	}
	return cats
}

func CategorySelectModal(categoryID int, categories []models.Category) g.Node {
	cats := createCategoryItems(categoryID, categories)

	return Div(
		ID("category-select-modal"),
		Class("fixed inset-0 bg-black/30 flex items-center justify-center z-50 p-8"),
		g.Attr("onclick", "this.remove()"),
		Div(
			Class("bg-white rounded-lg w-full shadow-2xl border-2 border-gray-300 flex flex-col"),
			Style("max-width: 400px; max-height: 80vh"),
			g.Attr("onclick", "event.stopPropagation()"),
			Div(
				Class("flex items-center justify-between p-6 border-b border-gray-200 flex-shrink-0"),
				H3(Class("text-xl font-bold text-gray-900"), g.Text("Select Category")),
				Button(
					Type("button"),
					Class("bg-white border-2 border-gray-800 rounded-full w-8 h-8 flex items-center justify-center shadow-lg hover:bg-gray-100 focus:outline-none cursor-pointer"),
					g.Attr("onclick", "this.closest('.fixed').remove()"),
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
					g.Group(cats),
				),
			),
		),
	)
}

func SearchContainerRefresh(categoryName, categoryImage string) g.Node {
	return g.Group([]g.Node{
		SearchContainer(categoryName, categoryImage),
		RemoveModalOOB("category-select-modal"),
	})
}

// RemoveModalOOB returns an out-of-band swap element to remove a modal by ID
func RemoveModalOOB(modalID string) g.Node {
	return Div(
		ID(modalID),
		hx.SwapOOB("outerHTML"),
	)
}

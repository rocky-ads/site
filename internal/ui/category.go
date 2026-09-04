package ui

import (
	"net/url"
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

const categorySelectOpenID = "category-select-open"

func CategorySelect(selected CategoryOption, categories []CategoryOption,
	returnParam string) g.Node {
	action := "Change category"
	items := make([]g.Node, len(categories))
	for i, cat := range categories {
		items[i] = CategoryItem(selected.ID, cat.ID, cat.Name,
			cat.ImageFile, returnParam)
	}
	imagePath := "/images/category/" + selected.ImageFile

	return Div(
		Class("relative inline-block z-50"),
		Input(
			Type("checkbox"),
			ID(categorySelectOpenID),
			Class("category-select-open"),
		),
		Label(
			For(categorySelectOpenID),
			Class("py-2 px-3 flex items-center gap-2 rounded-xl "+
				"border-2 border-blue-500 bg-blue-100 hover:bg-blue-200 "+
				"dark:bg-blue-900 dark:hover:bg-blue-800 "+
				"dark:border-blue-400 cursor-pointer text-left min-w-0"),
			g.Attr("aria-haspopup", "listbox"),
			g.Attr("aria-label", action+", currently "+selected.Name),
			g.Attr("title", action),
			Img(
				Src(imagePath),
				Alt(""),
				Class("w-6 h-6 shrink-0 dark:invert dark:opacity-80"),
			),
			Span(Class("font-medium truncate"), g.Text(selected.Name)),
			Img(
				Src("/images/expand.svg"),
				Alt(""),
				Class("w-5 h-5 shrink-0 opacity-70 dark:invert"),
			),
		),
		Label(
			For(categorySelectOpenID),
			Class("category-select-dismiss fixed inset-0 z-40"),
			g.Attr("aria-label", "Close category list"),
		),
		Div(
			Class("category-select-menu absolute left-0 z-50 mt-1 py-1 "+
				"rounded-xl border-2 border-zinc-300 dark:border-zinc-600 "+
				"bg-white dark:bg-zinc-800 shadow-lg overflow-y-auto "+
				"max-h-[85vh]"),
			g.Attr("role", "listbox"),
			g.Group(items),
		),
	)
}

func CategoryItem(currentCategoryID, categoryID int, name, imageFile,
	returnParam string) g.Node {
	itemClass := "flex items-center gap-2 px-3 py-2 hover:bg-zinc-50 " +
		"dark:hover:bg-zinc-800 cursor-pointer text-zinc-900 " +
		"dark:text-zinc-200 whitespace-nowrap"
	if categoryID == currentCategoryID {
		itemClass += " bg-blue-50 dark:bg-blue-900/30"
	}

	returnURL := "/api/category/" + strconv.Itoa(categoryID) +
		"/switch?return=" + url.QueryEscape(returnParam)

	attrs := []g.Node{
		Class(itemClass),
		g.Attr("role", "option"),
		hx.Get(returnURL),
		hx.Swap("none"),
	}
	if returnParam == "/" {
		attrs = append(attrs, hx.Include("#search-widget"))
	}
	attrs = append(attrs,
		Img(
			Src("/images/category/"+imageFile),
			Alt(""),
			Class("w-6 h-6 shrink-0 dark:invert dark:opacity-80"),
		),
		Span(g.Text(name)),
	)

	return Div(attrs...)
}

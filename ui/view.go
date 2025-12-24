package ui

import (
	"fmt"
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// View types
const (
	ViewGrid int = iota + 1
	ViewList
	ViewTree
)

var viewNames = map[int]string{
	ViewList: "list",
	ViewGrid: "grid",
	ViewTree: "tree",
}

func GetViewName(view int) string {
	return viewNames[view]
}

func ValidateView(viewStr string) int {
	view, err := strconv.Atoi(viewStr)
	if err != nil {
		return ViewGrid
	}
	if view < ViewList || view > ViewTree {
		return ViewGrid
	}
	return view
}

func viewToggle(view, target int) g.Node {
	active := view == target
	class := "p-2 rounded-full border-2 "
	if active {
		class += "border-blue-500 bg-blue-100 dark:bg-blue-900 dark:border-blue-400"
	} else {
		class += "border-transparent hover:bg-gray-100 dark:hover:bg-gray-800"
	}
	return Button(
		Type("button"),
		Class(class),
		hx.Get("/api/view/"+strconv.Itoa(target)),
		hx.Target("#search-view"),
		hx.Swap("outerHTML"),
		Img(
			Class("w-6 h-6 dark:invert"),
			Src("/images/"+GetViewName(target)+".svg"),
			Alt(GetViewName(target)+" view"),
		),
	)
}

func viewToggles(view int) g.Node {
	return Div(
		Class("flex items-center gap-2"),
		viewToggle(view, ViewGrid),
		viewToggle(view, ViewList),
		viewToggle(view, ViewTree),
	)
}

func AdGridNode(adID int, title string) g.Node {
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class("block"),
		ID(fmt.Sprintf("ad-%d", adID)),
		g.Text(title),
	)
}

func AdListNode(userID int, adID int, title string, active bool, bookmarked bool, csrfToken string) g.Node {
	class := "flex flex-wrap items-center justify-between py-2 px-3 cursor-pointer"
	if active {
		class += " hover:bg-gray-50 dark:hover:bg-gray-800"
	} else {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		ID(fmt.Sprintf("ad-%d", adID)),
		Div(
			Class("flex items-center gap-2 text-blue-600 hover:text-blue-800 min-w-0"),
			g.If(userID != 0, Bookmark(adID, bookmarked, csrfToken)),
			Span(Class("min-w-0"), g.Text(title)),
		),
	)
}

func AdTreeNode(adID int, title string) g.Node {
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class("block"),
		ID(fmt.Sprintf("ad-%d", adID)),
		g.Text(title),
	)
}

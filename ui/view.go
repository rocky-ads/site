package ui

import (
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// View types
const (
	ViewGrid int = iota + 1
	ViewList
	//ViewTree
)

var viewNames = map[int]string{
	ViewList: "list",
	ViewGrid: "grid",
}

func GetViewName(view int) string {
	return viewNames[view]
}

func ValidateView(viewStr string) int {
	view, err := strconv.Atoi(viewStr)
	if err != nil {
		return ViewGrid
	}
	if view < ViewGrid || view > ViewList {
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
		class += "border-transparent hover:bg-zinc-100 dark:hover:bg-zinc-800"
	}
	return Button(
		Type("button"),
		Class(class),
		hx.Get("/api/view/"+strconv.Itoa(target)),
		hx.Target("#search-view"),
		hx.Swap("outerHTML"),
		Img(
			Class("w-6 h-6 dark:invert dark:opacity-80"),
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
	)
}

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
	class := "p-1.5 rounded-full shrink-0 "
	if active {
		class += "bg-white dark:bg-zinc-600 shadow-sm"
	} else {
		class += "hover:bg-zinc-200/60 dark:hover:bg-zinc-700"
	}
	return Button(
		Type("button"),
		Class(class),
		g.Attr("aria-label", GetViewName(target)+" view"),
		g.Attr("aria-pressed", strconv.FormatBool(active)),
		hx.Get("/api/view/"+strconv.Itoa(target)),
		hx.Target("#search-view"),
		hx.Swap("outerHTML"),
		Img(
			Class("w-6 h-6 shrink-0 dark:invert dark:opacity-80"),
			Src("/images/"+GetViewName(target)+".svg"),
			Alt(""),
		),
	)
}

func viewToggles(view int) g.Node {
	return Div(
		Class("flex items-center shrink-0 rounded-full "+
			"bg-zinc-100 dark:bg-zinc-800 p-0.5"),
		viewToggle(view, ViewGrid),
		viewToggle(view, ViewList),
	)
}

package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

var hideModalOnClick = g.Attr("onclick",
	"document.querySelectorAll('.modal').forEach(el => el.classList.add('hidden'))")

func modalBackdrop() g.Node {
	return Div(
		ID("modal-backdrop"),
		hx.SwapOOB("true"),
		Class("modal fixed inset-0 bg-black/30 z-40"),
		hideModalOnClick,
	)
}

func modalPlaceholder() []g.Node {
	return []g.Node{
		Div(
			ID("modal-backdrop"),
			Class("modal hidden"),
		),
		Div(
			ID("modal"),
			Class("modal hidden"),
		),
	}
}

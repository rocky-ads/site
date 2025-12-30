package ui

import (
	"fmt"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// ErrorPageContent returns the content for error pages
func ErrorPageContent(code int, message string) []g.Node {
	return []g.Node{
		Div(
			Class("text-center py-16"),
			H1(
				Class("text-6xl font-bold mb-4"),
				g.Text(fmt.Sprintf("%d", code)),
			),
			H2(
				Class("text-2xl font-semibold mb-2"),
				g.Text(message),
			),
			P(
				Class("text-zinc-600 dark:text-zinc-400 mb-8"),
				g.Text("Sorry, something went wrong."),
			),
			standardButton(buttonProps{
				Text: "Go Home",
				Href: "/",
			}),
		),
	}
}

func ErrorDiv(errMsg string) g.Node {
	return Div(
		ID("error"),
		Class("text-red-500 text-sm"),
		g.If(errMsg != "", hx.SwapOOB("true")),
		g.If(errMsg != "", g.Text(errMsg)),
	)
}

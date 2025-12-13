package ui

import (
	"fmt"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ErrorPageContent returns the content for error pages
func ErrorPageContent(code int, message string) []g.Node {
	return []g.Node{
		Div(
			Class("text-center py-16"),
			H1(
				Class("text-6xl font-bold text-gray-900 dark:text-gray-100 mb-4"),
				g.Text(fmt.Sprintf("%d", code)),
			),
			H2(
				Class("text-2xl font-semibold text-gray-700 dark:text-gray-300 mb-2"),
				g.Text(message),
			),
			P(
				Class("text-gray-600 dark:text-gray-400 mb-8"),
				g.Text("Sorry, something went wrong."),
			),
			StandardButton(ButtonProps{
				Text: "Go Home",
				Href: "/",
			}),
		),
	}
}

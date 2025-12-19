package ui

import (
	g "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

// PageTitle creates a standardized H1 page title element
func PageTitle(text string) g.Node {
	return html.H1(
		html.Class("text-3xl font-bold mb-6"),
		g.Text(text),
	)
}

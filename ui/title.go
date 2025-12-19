package ui

import (
	g "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

// pageTitle creates a standardized H1 page title element
func pageTitle(text string) g.Node {
	return html.H1(
		html.Class("text-3xl font-bold mb-6"),
		g.Text(text),
	)
}

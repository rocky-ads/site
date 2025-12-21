package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func label(text string) g.Node {
	return Label(
		Class("block text-base font-medium mb-1"),
		g.Text(text),
	)
}


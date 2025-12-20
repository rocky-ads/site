package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func UserMenu(userName string, isAdmin bool) g.Node {
	return Div(
		Class("user-menu"),
		g.Text(userName),
	)
}

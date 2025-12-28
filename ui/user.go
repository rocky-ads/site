package ui

import (
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// NavigationIcon creates an icon for navigation menu items
func menuIcon(iconSrc, alt string) g.Node {
	return Img(
		Src(iconSrc),
		Alt(alt),
		Class("w-6 h-6 mr-2"),
	)
}

func menuHeader(userName string) g.Node {
	return Div(
		Class("px-4 py-3 border-b border-gray-100 text-center"),
		Div(
			Class("w-12 h-12 bg-red-500 text-white rounded-full flex items-center justify-center font-semibold text-lg mx-auto mb-2"),
			g.Text(getUserInitial(userName)),
		),
		Div(
			Class("text-sm font-medium text-gray-900"),
			g.Text(userName),
		),
		Div(
			Class("text-xs text-gray-500"),
			g.Text("Logged in"),
		),
	)
}

func UserMenu(userName string, isAdmin bool) g.Node {
	var menuItems []g.Node

	if isAdmin {
		menuItems = append(menuItems,
			A(
				Href("/admin/dashboard"),
				Class("block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center"),
				hideModalOnClick,
				menuIcon("/images/tools.svg", "Admin"),
				g.Text("Admin"),
			),
		)
	}

	menuItems = append(menuItems,
		A(
			Href("/ads"),
			Class("block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center"),
			hideModalOnClick,
			menuIcon("/images/bookmark-true.svg", "My Ads"),
			g.Text("My Ads"),
		),
		A(
			Href("/messages"),
			Class("block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center justify-between"),
			hideModalOnClick,
			Div(
				Class("flex items-center"),
				menuIcon("/images/message.svg", "Messages"),
				g.Text("Messages"),
			),
			Span(
				ID("menu-message-count"),
				Class("bg-green-500 text-white rounded-full h-6 min-w-6 px-1.5 flex items-center justify-center text-xs font-bold empty:hidden"),
				g.Attr("sse-swap", "message-count"),
				hx.Swap("innerHTML"),
				// Start empty - will be populated by SSE
			),
		),
		A(
			Href("/settings"),
			Class("block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center"),
			hideModalOnClick,
			menuIcon("/images/settings.svg", "Settings"),
			g.Text("Settings"),
		),
		A(
			Href("/about"),
			Class("block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center"),
			hideModalOnClick,
			menuIcon("/images/info.svg", "About"),
			g.Text("About"),
		),
		A(
			Href("/logout"),
			Class("block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center"),
			hideModalOnClick,
			menuIcon("/images/logout.svg", "Logout"),
			g.Text("Logout"),
		),
	)

	return g.Group([]g.Node{
		modalBackdrop(),
		Div(
			ID("modal"),
			hx.SwapOOB("true"),
			Class("modal fixed z-50 top-16 right-4"),
			Div(
				Class("bg-white rounded-lg shadow-lg border border-gray-200 w-40"),
				menuHeader(userName),
				Div(
					Class("py-1"),
					g.Group(menuItems),
				),
			),
		),
	})
}

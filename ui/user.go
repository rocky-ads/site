package ui

import (
	"fmt"

	"github.com/rocky-ads/site/config"
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

func menuHeader(userID int, userName string, eggCount int, userEggIcons g.Node) g.Node {
	return Div(
		Class("px-4 py-3 border-b border-zinc-100 text-center"),
		Div(
			Class("w-12 h-12 bg-red-500 text-white rounded-full flex items-center justify-center font-semibold text-lg mx-auto mb-2"),
			g.Text(getUserInitial(userName)),
		),
		Div(
			Class("flex items-center justify-center gap-2 mb-1"),
			Div(
				Class("text-sm font-medium text-zinc-900 flex items-center gap-1"),
				UserNameLink(userID, userName, userEggIcons),
			),
			EggCountBadge(eggCount),
		),
		Div(
			Class("text-xs text-zinc-500"),
			g.Text("Logged in"),
		),
	)
}

// UserNameLink renders a link to the user's profile with a hover popup summary
func UserNameLink(userID int, userName string, extra ...g.Node) g.Node {
	href := fmt.Sprintf("/auth/user/%d", userID)
	summaryURL := fmt.Sprintf("/auth/user/%d/summary", userID)
	popupID := fmt.Sprintf("user-summary-%d", userID)
	return Span(
		Class("relative group inline-flex items-center gap-1"),
		A(
			Href(href),
			Class("text-blue-600 dark:text-blue-400 hover:underline"),
			hx.Get(summaryURL),
			hx.Trigger("mouseenter delay:200ms once"),
			hx.Target("next .user-summary-popup"),
			hx.Swap("innerHTML"),
			g.Group(extra),
			g.Text(userName),
		),
		Span(
			Class("user-summary-popup absolute left-1/2 -translate-x-1/2 top-full mt-1 hidden group-hover:block z-50 w-64 p-3 bg-white dark:bg-zinc-800 rounded shadow-lg border border-zinc-200 dark:border-zinc-700 text-sm text-zinc-700 dark:text-zinc-300"),
			ID(popupID),
			g.Text("Loading…"),
		),
	)
}

func userProfileStats(d UserProfileData, memberSinceClass, statClass string) []g.Node {
	nodes := []g.Node{
		Div(Class(memberSinceClass), g.Text("Member since "+d.MemberSince)),
		Div(Class(statClass),
			g.Text(fmt.Sprintf("%d active ad(s)", d.ActiveAdCount)),
		),
	}
	if d.UserEggCount > 0 {
		nodes = append(nodes,
			Div(Class(statClass),
				g.Text(fmt.Sprintf("%d rock(s)", d.UserEggCount)),
			),
		)
	}
	return nodes
}

// UserSummaryFragment is the popup content loaded on hover
func UserSummaryFragment(d UserProfileData) g.Node {
	statClass := "text-xs text-zinc-500"
	return Div(
		Class("space-y-1"),
		Div(Class("font-semibold text-zinc-900 dark:text-zinc-100"), g.Text(d.Name)),
		g.Group(userProfileStats(d, statClass, statClass)),
	)
}

// UserProfilePage renders the user profile page body
func UserProfilePage(d UserProfileData) []g.Node {
	return []g.Node{
		pageTitle(d.Name),
		Div(
			Class("mt-8 space-y-4 max-w-md"),
			Div(
				Class("flex items-center gap-3"),
				Div(
					Class("w-16 h-16 bg-red-500 text-white rounded-full flex items-center justify-center font-semibold text-2xl"),
					g.Text(getUserInitial(d.Name)),
				),
				Div(
					Class("text-zinc-600 dark:text-zinc-400"),
					g.Group(userProfileStats(d, "text-sm", "text-sm mt-1")),
				),
			),
		),
	}
}

func UserMenu(userName string, userID int, isAdmin bool, hasUnread bool, eggCount int, userEggCount int) g.Node {
	var menuItems []g.Node

	if isAdmin {
		menuItems = append(menuItems,
			A(
				Href("/admin/dashboard"),
				Class("block px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50 flex items-center"),
				menuIcon("/images/tools.svg", "Admin"),
				g.Text("Admin"),
			),
		)
	}

	menuItems = append(menuItems,
		A(
			Href("/auth/user/myads"),
			Class("block px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50 flex items-center"),
			menuIcon("/images/bookmark-true.svg", "My Ads"),
			g.Text("My Ads"),
		),
		A(
			Href("/auth/user/messages"),
			Class("block px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50 flex items-center justify-between relative"),
			Div(
				Class("flex items-center"),
				menuIcon("/images/message.svg", "Messages"),
				g.Text("Messages"),
			),
			Div(
				ID("message-unread-indicator"),
				g.If(hasUnread, hasUnreadIndicator()),
			),
		),
		A(
			Href("/auth/user/settings"),
			Class("block px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50 flex items-center"),
			menuIcon("/images/settings.svg", "Settings"),
			g.Text("Settings"),
		),
		A(
			Href("/auth/user/about"),
			Class("block px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50 flex items-center"),
			menuIcon("/images/info.svg", "About"),
			g.Text("About"),
		),
		A(
			Href("/logout"),
			Class("block px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50 flex items-center"),
			menuIcon("/images/logout.svg", "Logout"),
			g.Text("Logout"),
		),
	)

	return g.Group([]g.Node{
		modalBackdrop("user-menu"),
		Div(
			ID("user-menu-modal"),
			Class("fixed z-50 top-20 right-6"),
			Div(
				Class("bg-white rounded-lg shadow-lg border border-zinc-200 w-40"),
				menuHeader(userID, userName, eggCount, UserEggIcons(userID, userEggCount)),
				Div(
					Class("py-1"),
					g.Group(menuItems),
				),
			),
		),
	})
}

func MyAdsPage(activeTab string, adNodes []g.Node) []g.Node {
	return []g.Node{
		pageTitle("My Ads"),
		MyAdsContainer(activeTab, adNodes),
	}
}

func MyAdsContainer(activeTab string, adNodes []g.Node) g.Node {
	return Div(
		ID("my-ads-container"),
		Class("space-y-6 mt-6"),
		MyAdsTabs(activeTab),
		MyAdsContent(activeTab, adNodes),
	)
}

func MyAdsTabs(activeTab string) g.Node {
	return Div(
		ID("my-ads-tabs"),
		Class("border-b border-zinc-200 dark:border-zinc-700"),
		Div(
			Class("flex space-x-8"),
			myAdsTab("Bookmarked", "bookmarked", activeTab == "bookmarked"),
			myAdsTab("Active", "active", activeTab == "active"),
			myAdsTab("Deleted", "deleted", activeTab == "deleted"),
		),
	)
}

func myAdsTab(name, tabID string, active bool) g.Node {
	var classes string
	if active {
		classes = "border-b-2 border-blue-500 text-blue-600 dark:text-blue-400 py-4 px-1 text-sm font-medium"
	} else {
		classes = "border-b-2 border-transparent text-zinc-500 hover:text-zinc-700 hover:border-zinc-300 dark:text-zinc-400 dark:hover:text-zinc-300 py-4 px-1 text-sm font-medium"
	}

	href := fmt.Sprintf("/auth/user/myads/tab/%s", tabID)

	return A(
		Href(href),
		hx.Get(href),
		hx.Target("#my-ads-container"),
		hx.Swap("outerHTML"),
		Class(classes),
		g.Text(name),
	)
}

func MyAdsContent(activeTab string, adNodes []g.Node) g.Node {
	return Div(
		ID("my-ads-content"),
		Class("mt-4"),
		g.If(len(adNodes) == 0,
			Div(
				Class("text-center text-zinc-600 dark:text-zinc-400 py-8"),
				P(g.Text("No ads found.")),
			),
		),
		g.If(len(adNodes) > 0,
			Div(
				Class("space-y-0"),
				g.Group(adNodes),
			),
		),
	)
}

func SettingsPage() []g.Node {
	return []g.Node{
		pageTitle("Settings"),
		Div(
			Class("mt-8 text-center text-zinc-600 dark:text-zinc-400"),
			P(g.Text("This page will show your settings.")),
		),
	}
}

func AboutPage() []g.Node {
	return []g.Node{
		pageTitle("About"),
		Div(
			Class("mt-8 text-center text-zinc-600 dark:text-zinc-400"),
			P(g.Textf("This page will show information about %s.", config.ServerName)),
		),
	}
}

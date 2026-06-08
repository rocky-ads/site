package ui

import (
	"fmt"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/password"
	"github.com/rocky-ads/site/internal/phoneformat"
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

func menuHeader(userID int, userName, memberSince string, eggCount int, userEggIcons g.Node) g.Node {
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
			Class("text-[10px] text-zinc-500"),
			g.Text("Member since "+memberSince),
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

func UserMenu(userName, memberSince string, userID int, isAdmin bool, hasUnread bool, eggCount int, userEggCount int) g.Node {
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
				menuHeader(userID, userName, memberSince, eggCount, UserEggIcons(userID, userEggCount)),
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

const (
	SettingsChangePasswordErrorID = "change-password-error"
	SettingsDeleteAccountErrorID  = "delete-account-error"
)

func settingsFormActions(button buttonProps, errorID string) g.Node {
	return Div(
		Class("flex items-center gap-4"),
		standardButton(button),
		ErrorDivWithID(errorID, ""),
	)
}

func settingsSection(title string, children ...g.Node) g.Node {
	nodes := []g.Node{
		H2(Class("text-xl font-semibold mb-4"), g.Text(title)),
	}
	nodes = append(nodes, children...)
	return Div(
		Class("mt-8 p-6 border border-zinc-200 dark:border-zinc-700 rounded-lg"),
		g.Group(nodes),
	)
}

func settingsAccountRow(labelText string, value g.Node) g.Node {
	return Div(
		Class("flex items-center gap-6"),
		Span(
			Class("w-24 shrink-0 text-sm font-medium text-zinc-500 dark:text-zinc-400"),
			g.Text(labelText),
		),
		value,
	)
}

func settingsUserAvatar(userName string) g.Node {
	return Span(
		Class("bg-red-500 text-white rounded-full w-8 h-8 flex items-center justify-center font-semibold text-sm shrink-0"),
		g.Text(getUserInitial(userName)),
	)
}

func settingsAccountSection(userName, phoneE64 string) g.Node {
	return settingsSection("Account",
		Div(
			Class("space-y-4"),
			settingsAccountRow("Username",
				Div(
					Class("flex items-center gap-2 min-w-0"),
					settingsUserAvatar(userName),
					Span(
						Class("text-zinc-900 dark:text-zinc-100 font-medium"),
						g.Text(userName),
					),
				),
			),
			settingsAccountRow("Phone",
				Span(
					Class("text-zinc-900 dark:text-zinc-100"),
					g.Text(phoneformat.Display(phoneE64)),
				),
			),
		),
	)
}

func namedPasswordInput(name, autocomplete string) g.Node {
	return Input(
		Class("w-full p-2 border rounded-md dark:bg-zinc-800 dark:border-zinc-600"),
		Type("password"),
		Name(name),
		MaxLength("32"),
		g.Attr("autocomplete", autocomplete),
		Required(),
	)
}

func NotificationsSection(smsOptedOut bool) g.Node {
	enabled := !smsOptedOut
	statusClass := "text-green-600 dark:text-green-400"
	statusText := "ON"
	if smsOptedOut {
		statusClass = "text-red-600 dark:text-red-400"
		statusText = "OFF"
	}

	return Div(
		ID("sms-notifications"),
		Class("mt-8 p-6 border border-zinc-200 dark:border-zinc-700 rounded-lg"),
		H2(Class("text-xl font-semibold mb-4"), g.Text("Notifications")),
		Div(
			Class("flex items-center justify-between gap-4"),
			Div(
				Class("flex items-center gap-3"),
				Label(
					Class("relative inline-flex items-center cursor-pointer"),
					Input(
						Type("checkbox"),
						Name("enabled"),
						Value("true"),
						Class("sr-only peer"),
						g.If(enabled, Checked()),
						hx.Post("/auth/user/settings/notifications"),
						hx.Target("#sms-notifications"),
						hx.Swap("outerHTML"),
						hx.Trigger("change"),
					),
					Span(
						Class("w-11 h-6 bg-zinc-300 dark:bg-zinc-600 rounded-full peer peer-checked:bg-green-500 peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-0.5 after:left-0.5 after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all"),
					),
				),
				Span(Class("text-sm font-medium"), g.Text("Text messages")),
			),
			Span(
				Class("text-sm font-semibold "+statusClass),
				g.Text("Text messages: "+statusText),
			),
		),
		g.If(enabled,
			P(
				Class("mt-4 text-sm"),
				A(
					Href("/auth/user/about#sms-notifications"),
					Class("text-blue-600 dark:text-blue-400 hover:underline"),
					g.Text("Why am I not getting text messages?"),
				),
			),
		),
	)
}

func SettingsPage(name, phoneE64 string, smsOptedOut bool) []g.Node {
	return []g.Node{
		pageTitle("Settings"),
		settingsAccountSection(name, phoneE64),
		NotificationsSection(smsOptedOut),
		settingsSection("Change Password",
			Form(
				Class("space-y-4"),
				hx.Post("/auth/user/settings/password"),
				hx.Swap("none"),
				Div(
					label("Current Password"),
					namedPasswordInput("current_password", "current-password"),
				),
				Div(
					label("New Password"),
					namedPasswordInput("new_password", "new-password"),
					Span(
						Class("text-xs text-zinc-500 dark:text-zinc-400 mt-1 block"),
						g.Text(password.StrengthRequirements),
					),
				),
				Div(
					label("Confirm New Password"),
					namedPasswordInput("confirm_password", "new-password"),
				),
				settingsFormActions(buttonProps{
					Type: "submit",
					Text: "Change Password",
				}, SettingsChangePasswordErrorID),
			),
		),
		settingsSection("Delete Account",
			P(
				Class("text-sm text-zinc-600 dark:text-zinc-400 mb-4"),
				g.Text("Permanently delete your account. This action cannot be undone."),
			),
			Form(
				Class("space-y-4"),
				hx.Post("/auth/user/settings/delete"),
				hx.Swap("none"),
				Div(
					label("Password"),
					namedPasswordInput("password", "current-password"),
				),
				settingsFormActions(buttonProps{
					Type:  "submit",
					Text:  "Delete My Account",
					Class: "bg-red-600 hover:bg-red-700",
					Attrs: []g.Node{
						hx.Confirm("Are you sure you want to permanently delete your account? This cannot be undone."),
					},
				}, SettingsDeleteAccountErrorID),
			),
		),
	}
}

func AboutPage() []g.Node {
	fromNumber := config.TwilioFromNumber
	smsResumeHelp := "If you previously replied STOP to a text message from us, your carrier may have blocked further texts. Text START to our number to resume delivery."
	if fromNumber != "" {
		smsResumeHelp = fmt.Sprintf(
			"If you previously replied STOP to a text message from us, your carrier may have blocked further texts. Text START to %s to resume delivery.",
			fromNumber,
		)
	}

	return []g.Node{
		pageTitle("About"),
		Div(
			Class("mt-8 space-y-8 text-zinc-700 dark:text-zinc-300"),
			P(g.Textf("Welcome to %s.", config.ServerName)),
			Div(
				ID("sms-notifications"),
				H2(Class("text-xl font-semibold mb-3"), g.Text("FAQ")),
				Div(
					Class("space-y-2"),
					H3(
						Class("text-lg font-medium"),
						g.Text("Why am I not getting text messages?"),
					),
					P(
						Class("text-sm text-zinc-600 dark:text-zinc-400"),
						g.Text("First, open Settings and confirm text messages are turned on."),
					),
					P(
						Class("text-sm text-zinc-600 dark:text-zinc-400"),
						g.Text(smsResumeHelp),
					),
					P(
						Class("text-sm text-zinc-600 dark:text-zinc-400"),
						g.Text("You can also turn notifications off or on anytime from Settings without texting STOP or START."),
					),
				),
			),
		),
	}
}

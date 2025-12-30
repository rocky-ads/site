package ui

import (
	"fmt"
	"strconv"

	"github.com/rocky-ads/site/user"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func AdminDashboardPage(users []user.User, sortBy, sortOrder string, currentUserID int) []g.Node {
	return []g.Node{
		pageTitle("Admin Dashboard"),
		AdminDashboardContainer("users", users, sortBy, sortOrder, currentUserID),
	}
}

func AdminDashboardContainer(activeTab string, users []user.User, sortBy, sortOrder string, currentUserID int) g.Node {
	return Div(
		ID("admin-dashboard-container"),
		Class("space-y-6 mt-6"),
		AdminTabs(activeTab),
		AdminContent(activeTab, users, sortBy, sortOrder, currentUserID),
	)
}

func AdminTabs(activeTab string) g.Node {
	return Div(
		ID("admin-tabs"),
		Class("border-b border-zinc-200 dark:border-zinc-700"),
		Div(
			Class("flex space-x-8"),
			adminTab("Users", "users", activeTab == "users"),
			adminTab("Settings", "settings", activeTab == "settings"),
		),
	)
}

func adminTab(name, tabID string, active bool) g.Node {
	var classes string
	if active {
		classes = "border-b-2 border-blue-500 text-blue-600 dark:text-blue-400 py-4 px-1 text-sm font-medium"
	} else {
		classes = "border-b-2 border-transparent text-zinc-500 hover:text-zinc-700 hover:border-zinc-300 dark:text-zinc-400 dark:hover:text-zinc-300 py-4 px-1 text-sm font-medium"
	}

	href := fmt.Sprintf("/admin/tab/%s", tabID)

	return A(
		Href(href),
		hx.Get(href),
		hx.Target("#admin-dashboard-container"),
		hx.Swap("outerHTML"),
		Class(classes),
		g.Text(name),
	)
}

func AdminContent(activeTab string, users []user.User, sortBy, sortOrder string, currentUserID int) g.Node {
	return Div(
		ID("admin-content"),
		g.If(activeTab == "users", Div(
			Class("mt-4"),
			UsersTable(users, sortBy, sortOrder, currentUserID),
		)),
		g.If(activeTab == "settings", AdminSettingsTab()),
	)
}

func UsersRows(users []user.User, currentUserID int) g.Node {
	return Div(
		ID("users-rows"),
		Class("bg-white dark:bg-zinc-900 divide-y divide-zinc-200 dark:divide-zinc-700"),
		g.Group(userRows(users, currentUserID)),
	)
}

func AdminSettingsTab() g.Node {
	return Div(
		Class("mt-4 p-6 bg-white dark:bg-zinc-800 rounded-lg shadow"),
		H2(
			Class("text-xl font-semibold text-zinc-900 dark:text-zinc-200 mb-4"),
			g.Text("Settings"),
		),
		P(
			Class("text-zinc-600 dark:text-zinc-400"),
			g.Text("This is a dummy settings tab for testing purposes."),
		),
	)
}

func UsersTable(users []user.User, sortBy, sortOrder string, currentUserID int) g.Node {
	return Div(
		ID("users-table"),
		Class("w-full text-xs"),
		tableHeader(sortBy, sortOrder),
		Div(
			ID("users-rows"),
			Class("bg-white dark:bg-zinc-900 divide-y divide-zinc-200 dark:divide-zinc-700"),
			g.Group(userRows(users, currentUserID)),
		),
	)
}

func tableHeader(sortBy, sortOrder string) g.Node {
	return Div(
		Class("grid grid-cols-7 gap-2 bg-zinc-50 dark:bg-zinc-800 px-2 py-2 border-b border-zinc-200 dark:border-zinc-700"),
		sortableHeader("id", "ID", sortBy, sortOrder),
		sortableHeader("name", "Name", sortBy, sortOrder),
		sortableHeader("phone", "Phone", sortBy, sortOrder),
		sortableHeader("is_admin", "Admin", sortBy, sortOrder),
		sortableHeader("created_at", "Created", sortBy, sortOrder),
		sortableHeader("deleted_at", "Deleted", sortBy, sortOrder),
		Div(
			Class("px-2 py-2 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider"),
			g.Text("Actions"),
		),
	)
}

func sortableHeader(column, label, currentSort, currentOrder string) g.Node {
	var href string
	var order string

	if currentSort == column {
		if currentOrder == "ASC" {
			order = "DESC"
		} else {
			order = "ASC"
		}
	} else {
		order = "ASC"
	}

	href = fmt.Sprintf("/admin/users?sort=%s&order=%s", column, order)

	var sortIndicator g.Node
	if currentSort == column {
		if currentOrder == "ASC" {
			sortIndicator = Span(g.Text(" ↑"))
		} else {
			sortIndicator = Span(g.Text(" ↓"))
		}
	}

	return Div(
		Class("px-2 py-2 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider"),
		A(
			Href(href),
			hx.Get(href),
			hx.Target("#users-table"),
			hx.Swap("outerHTML"),
			Class("flex items-center cursor-pointer hover:text-zinc-700 dark:hover:text-zinc-300"),
			g.Text(label),
			sortIndicator,
		),
	)
}

func userRows(users []user.User, currentUserID int) []g.Node {
	rows := make([]g.Node, len(users))
	for i, u := range users {
		rows[i] = UserRow(u, currentUserID)
	}
	return rows
}

func UserRow(u user.User, currentUserID int) g.Node {
	isDeleted := u.DeletedAt != nil
	adminStatus := "No"
	if u.IsAdmin {
		adminStatus = "Yes"
	}

	var statusClass string
	var rowClass string

	if isDeleted {
		statusClass = "text-zinc-400 dark:text-zinc-600"
		rowClass = ""
	} else {
		statusClass = "text-zinc-900 dark:text-zinc-200"
		if u.IsAdmin {
			rowClass = "bg-red-50 dark:bg-red-900/20"
		} else {
			rowClass = ""
		}
	}

	createdAtStr := u.CreatedAt.Format("2006-01-02 15:04:05")
	var deletedAtStr string
	if u.DeletedAt != nil {
		deletedAtStr = u.DeletedAt.Format("2006-01-02 15:04:05")
	} else {
		deletedAtStr = "—"
	}

	rowID := fmt.Sprintf("user-row-%d", u.ID)

	return Div(
		ID(rowID),
		Class("grid grid-cols-7 gap-2 px-2 py-2 border-b border-zinc-200 dark:border-zinc-700 "+rowClass),
		Div(
			Class("px-2 py-2 text-xs "+statusClass),
			g.Text(strconv.Itoa(u.ID)),
		),
		Div(
			Class("px-2 py-2 text-xs "+statusClass),
			g.Text(u.Name),
		),
		Div(
			Class("px-2 py-2 text-xs "+statusClass),
			g.Text(u.PhoneE64),
		),
		Div(
			Class("px-2 py-2 text-xs "+statusClass),
			g.Text(adminStatus),
		),
		Div(
			Class("px-2 py-2 text-xs "+statusClass),
			g.Text(createdAtStr),
		),
		Div(
			Class("px-2 py-2 text-xs "+statusClass),
			g.Text(deletedAtStr),
		),
		Div(
			Class("px-2 py-2 text-xs "+statusClass),
			userActions(u, currentUserID),
		),
	)
}

func userActions(u user.User, currentUserID int) g.Node {
	var actions []g.Node
	rowID := fmt.Sprintf("user-row-%d", u.ID)

	if u.DeletedAt != nil {
		actions = append(actions,
			actionIconButton(
				"/images/restore.svg",
				"Restore user",
				fmt.Sprintf("/admin/user/%d/restore", u.ID),
				fmt.Sprintf("Are you sure you want to restore user %s?", u.Name),
				"text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300",
				rowID,
			),
		)
	} else {
		actions = append(actions,
			actionIconButton(
				"/images/trashcan.svg",
				"Delete user",
				fmt.Sprintf("/admin/user/%d/delete", u.ID),
				fmt.Sprintf("Are you sure you want to delete user %s?", u.Name),
				"text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300",
				rowID,
			),
		)

		if u.IsAdmin {
			// Don't show demote button if user is trying to demote themselves
			if u.ID != currentUserID {
				actions = append(actions,
					actionIconButton(
						"/images/demote.svg",
						"Demote from admin",
						fmt.Sprintf("/admin/user/%d/demote", u.ID),
						fmt.Sprintf("Are you sure you want to demote user %s from admin?", u.Name),
						"text-yellow-600 hover:text-yellow-900 dark:text-yellow-400 dark:hover:text-yellow-300",
						rowID,
					),
				)
			}
		} else {
			actions = append(actions,
				actionIconButton(
					"/images/promote.svg",
					"Promote to admin",
					fmt.Sprintf("/admin/user/%d/promote", u.ID),
					fmt.Sprintf("Are you sure you want to promote user %s to admin?", u.Name),
					"text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300",
					rowID,
				),
			)
		}
	}

	return Div(
		Class("flex items-center flex-wrap gap-1"),
		g.Group(actions),
	)
}

func actionIconButton(iconSrc, alt, actionURL, confirmMsg, colorClass string, targetRowID string) g.Node {
	return Button(
		Type("button"),
		Class("p-1 rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors "+colorClass),
		hx.Post(actionURL),
		hx.Target(fmt.Sprintf("#%s", targetRowID)),
		hx.Swap("outerHTML"),
		g.Attr("hx-confirm", confirmMsg),
		Title(alt),
		Img(
			Src(iconSrc),
			Alt(alt),
			Class("w-5 h-5 dark:invert dark:opacity-80"),
		),
	)
}

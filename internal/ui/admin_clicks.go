package ui

import (
	"fmt"
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func ClicksTab(d ClickAdminData) g.Node {
	return Div(
		ID("clicks-tab"),
		Class("mt-4 space-y-6"),
		hx.Get("/admin/tab/clicks"),
		hx.Target("#admin-dashboard-container"),
		hx.Swap("outerHTML"),
		hx.Trigger("every 10s"),
		P(
			Class("text-sm text-zinc-600 dark:text-zinc-400"),
			g.Text("Logged-in user engagement only. Anonymous ad views are not recorded."),
		),
		clickSummaryCards(d),
		clickTopAdsSection(d.TopAds),
		clickTopImagesSection(d.TopImages),
	)
}

func clickSummaryCards(d ClickAdminData) g.Node {
	return Div(
		Class("grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4"),
		clickStatCard("Users", strconv.Itoa(d.UsersWithClicks),
			"bg-violet-50 dark:bg-violet-900/20 border-violet-200 dark:border-violet-800",
			"text-violet-800 dark:text-violet-200", "text-violet-900 dark:text-violet-100"),
		clickStatCard("Ads clicked", strconv.Itoa(d.AdsClicked),
			"bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800",
			"text-blue-800 dark:text-blue-200", "text-blue-900 dark:text-blue-100"),
		clickStatCard("Ad detail views", strconv.Itoa(d.AdDetailViews),
			"bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800",
			"text-green-800 dark:text-green-200", "text-green-900 dark:text-green-100"),
		clickStatCard("Image navigations", strconv.Itoa(d.ImageNavClicks),
			"bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800",
			"text-amber-800 dark:text-amber-200", "text-amber-900 dark:text-amber-100"),
		clickStatCard("Active (7 days)", strconv.Itoa(d.ActiveLast7Days),
			"bg-zinc-50 dark:bg-zinc-800/50 border-zinc-200 dark:border-zinc-700",
			"text-zinc-700 dark:text-zinc-300", "text-zinc-900 dark:text-zinc-100"),
	)
}

func clickStatCard(label, value, boxClass, labelClass, valueClass string) g.Node {
	return Div(
		Class("border rounded-lg p-4 "+boxClass),
		Div(Class("text-sm font-medium "+labelClass), g.Text(label)),
		Div(Class("text-2xl font-bold mt-1 "+valueClass), g.Text(value)),
	)
}

func clickTopAdsSection(rows []ClickAdRow) g.Node {
	return clickTableSection(
		"Top ads by engagement",
		"Ads ranked by logged-in ad views plus image navigations.",
		clickTopAdHeader(),
		clickTopAdRows(rows),
	)
}

func clickTopAdRows(rows []ClickAdRow) []g.Node {
	if len(rows) == 0 {
		return nil
	}
	out := make([]g.Node, len(rows))
	for i, r := range rows {
		out[i] = Div(
			Class("grid grid-cols-6 gap-2 px-4 py-2 text-xs text-zinc-900 dark:text-zinc-200"),
			Div(
				Class("truncate"),
				A(
					Href(fmt.Sprintf("/ad/%d", r.AdID)),
					Class("text-blue-600 dark:text-blue-400 hover:underline"),
					g.Text(r.Title),
				),
			),
			Div(Class("truncate"), g.Text(r.CategoryName)),
			Div(g.Text(strconv.Itoa(r.UserCount))),
			Div(g.Text(strconv.Itoa(r.AdViews))),
			Div(g.Text(strconv.Itoa(r.ImageClicks))),
			Div(g.Text(r.LastActivity)),
		)
	}
	return out
}

func clickTopImagesSection(rows []ClickImageRow) g.Node {
	return clickTableSection(
		"Top images",
		"Carousel images with the most logged-in navigations.",
		clickTopImageHeader(),
		clickTopImageRows(rows),
	)
}

func clickTopImageRows(rows []ClickImageRow) []g.Node {
	if len(rows) == 0 {
		return nil
	}
	out := make([]g.Node, len(rows))
	for i, r := range rows {
		out[i] = Div(
			Class("grid grid-cols-5 gap-2 px-4 py-2 text-xs text-zinc-900 dark:text-zinc-200"),
			Div(
				Class("truncate"),
				A(
					Href(fmt.Sprintf("/ad/%d", r.AdID)),
					Class("text-blue-600 dark:text-blue-400 hover:underline"),
					g.Text(r.Title),
				),
			),
			Div(g.Text(strconv.Itoa(r.ImageIndex))),
			Div(g.Text(strconv.Itoa(r.UserCount))),
			Div(g.Text(strconv.Itoa(r.Clicks))),
			Div(g.Text(r.LastClick)),
		)
	}
	return out
}

func clickTopAdHeader() g.Node {
	return Div(
		Class("grid grid-cols-6 gap-2 bg-zinc-50 dark:bg-zinc-900 px-4 py-2 border-b border-zinc-200 dark:border-zinc-700 text-xs"),
		clickHeaderCell("Ad"),
		clickHeaderCell("Category"),
		clickHeaderCell("Users"),
		clickHeaderCell("Ad views"),
		clickHeaderCell("Image clicks"),
		clickHeaderCell("Last activity"),
	)
}

func clickTopImageHeader() g.Node {
	return Div(
		Class("grid grid-cols-5 gap-2 bg-zinc-50 dark:bg-zinc-900 px-4 py-2 border-b border-zinc-200 dark:border-zinc-700 text-xs"),
		clickHeaderCell("Ad"),
		clickHeaderCell("Image"),
		clickHeaderCell("Users"),
		clickHeaderCell("Clicks"),
		clickHeaderCell("Last clicked"),
	)
}

func clickHeaderCell(label string) g.Node {
	return Div(
		Class("font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider"),
		g.Text(label),
	)
}

func clickTableSection(title, subtitle string, header g.Node,
	rows []g.Node) g.Node {
	var body g.Node
	if len(rows) == 0 {
		body = Div(
			Class("px-4 py-6 text-sm text-zinc-500 dark:text-zinc-400"),
			g.Text("No click data yet."),
		)
	} else {
		body = Div(
			Class("divide-y divide-zinc-200 dark:divide-zinc-700"),
			g.Group(rows),
		)
	}
	return Div(
		Class("bg-white dark:bg-zinc-800 rounded-lg shadow overflow-hidden border border-zinc-200 dark:border-zinc-700"),
		Div(
			Class("px-4 py-3 border-b border-zinc-200 dark:border-zinc-700"),
			H2(
				Class("text-lg font-semibold text-zinc-900 dark:text-zinc-200"),
				g.Text(title),
			),
			P(
				Class("text-sm text-zinc-600 dark:text-zinc-400 mt-1"),
				g.Text(subtitle),
			),
		),
		header,
		body,
	)
}

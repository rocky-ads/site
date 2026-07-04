package ui

import (
	"fmt"
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func EmbeddingsTab(d EmbeddingAdminData) g.Node {
	refreshURL := fmt.Sprintf("/admin/tab/embeddings?category=%d", d.CategoryID)
	return Div(
		ID("embeddings-tab"),
		Class("mt-4 space-y-6"),
		hx.Get(refreshURL),
		hx.Target("#admin-dashboard-container"),
		hx.Swap("outerHTML"),
		hx.Trigger("every 10s"),
		embeddingSummaryCards(d),
		embeddingProviderRow(d),
		embeddingActions(),
		embeddingCacheSection(d.Caches),
		embeddingActivitiesSection(d),
		embeddingMissingSection(d.MissingAds),
	)
}

func embeddingProviderRow(d EmbeddingAdminData) g.Node {
	name := d.EmbedderProvider
	if name == "" {
		name = "unknown"
	}
	if d.EmbedderModel != "" {
		name += " · " + d.EmbedderModel
	}
	return Div(
		Class("flex items-center gap-2"),
		Span(
			Class("text-xs font-medium uppercase tracking-wide text-zinc-500 dark:text-zinc-400"),
			g.Text("Embedder"),
		),
		Span(
			Class("inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium "+
				"bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 "+
				"border border-indigo-200 dark:border-indigo-800"),
			g.Text(name),
		),
	)
}

func embeddingSummaryCards(d EmbeddingAdminData) g.Node {
	return Div(
		Class("grid grid-cols-1 sm:grid-cols-3 gap-4"),
		embeddingStatCard(
			"Embedded ads",
			strconv.Itoa(d.EmbeddedCount),
			"bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800",
			"text-green-800 dark:text-green-200",
			"text-green-900 dark:text-green-100",
		),
		embeddingStatCard(
			"Missing embeddings",
			strconv.Itoa(d.MissingCount),
			"bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800",
			"text-amber-800 dark:text-amber-200",
			"text-amber-900 dark:text-amber-100",
		),
		embeddingStatCard(
			"Queue depth",
			strconv.Itoa(d.QueueDepth),
			"bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800",
			"text-blue-800 dark:text-blue-200",
			"text-blue-900 dark:text-blue-100",
		),
	)
}

func embeddingStatCard(label, value, boxClass, labelClass,
	valueClass string) g.Node {
	return Div(
		Class("border rounded-lg p-4 "+boxClass),
		Div(Class("text-sm font-medium "+labelClass), g.Text(label)),
		Div(Class("text-2xl font-bold mt-1 "+valueClass), g.Text(value)),
	)
}

func embeddingActions() g.Node {
	return Div(
		Class("flex flex-wrap gap-3"),
		embeddingActionButton(
			"Backfill missing",
			"/admin/embeddings/backfill",
			"Start background embedding for ads without vectors",
		),
		embeddingActionButton(
			"Clear query cache",
			"/admin/embeddings/cache/query/clear",
			"Clear cached search-query embeddings",
		),
		embeddingActionButton(
			"Clear user cache",
			"/admin/embeddings/cache/user/clear",
			"Clear cached per-user search embeddings",
		),
		embeddingActionButton(
			"Clear site cache",
			"/admin/embeddings/cache/site/clear",
			"Clear cached site-level search embeddings",
		),
		embeddingActionButton(
			"Clear all caches",
			"/admin/embeddings/cache/clear-all",
			"Clear query, user, and site embedding caches",
		),
	)
}

func embeddingActionButton(label, url, confirm string) g.Node {
	return Button(
		Type("button"),
		Class("px-4 py-2 text-sm font-medium rounded-md border border-zinc-300 dark:border-zinc-600 "+
			"bg-white dark:bg-zinc-800 text-zinc-800 dark:text-zinc-200 "+
			"hover:bg-zinc-50 dark:hover:bg-zinc-700"),
		hx.Post(url),
		hx.Target("#admin-dashboard-container"),
		hx.Swap("outerHTML"),
		g.Attr("hx-confirm", confirm),
		g.Text(label),
	)
}

func embeddingCacheSection(caches []EmbeddingCachePanel) g.Node {
	panels := make([]g.Node, len(caches))
	for i, c := range caches {
		panels[i] = embeddingCachePanel(c)
	}
	return Div(
		Class("space-y-4"),
		H2(
			Class("text-lg font-semibold text-zinc-900 dark:text-zinc-200"),
			g.Text("Search embedding caches"),
		),
		P(
			Class("text-sm text-zinc-600 dark:text-zinc-400"),
			g.Text("In-memory caches for query, user, and site vectors used at search time."),
		),
		g.Group(panels),
	)
}

func embeddingCachePanel(c EmbeddingCachePanel) g.Node {
	return Div(
		Class("bg-white dark:bg-zinc-800 rounded-lg shadow border border-zinc-200 dark:border-zinc-700 p-4"),
		H3(
			Class("text-sm font-semibold text-zinc-900 dark:text-zinc-200 mb-3"),
			g.Text(c.Name),
		),
		Div(
			Class("grid grid-cols-2 sm:grid-cols-5 gap-3 text-sm"),
			embeddingCacheMetric("Hits", strconv.FormatInt(c.Hits, 10)),
			embeddingCacheMetric("Misses", strconv.FormatInt(c.Misses, 10)),
			embeddingCacheMetric("Hit rate", fmt.Sprintf("%.1f%%", c.HitRatePct)),
			embeddingCacheMetric("Items", strconv.FormatInt(c.ItemCount, 10)),
			embeddingCacheMetric("Memory", fmt.Sprintf("%.0f KB", c.MemoryKB)),
		),
	)
}

func embeddingCacheMetric(label, value string) g.Node {
	return Div(
		Div(
			Class("text-xs text-zinc-500 dark:text-zinc-400 uppercase tracking-wide"),
			g.Text(label),
		),
		Div(
			Class("font-medium text-zinc-900 dark:text-zinc-200"),
			g.Text(value),
		),
	)
}

func embeddingMissingSection(rows []MissingEmbeddingRow) g.Node {
	return Div(
		Class("bg-white dark:bg-zinc-800 rounded-lg shadow overflow-hidden border border-zinc-200 dark:border-zinc-700"),
		Div(
			Class("px-4 py-3 border-b border-zinc-200 dark:border-zinc-700"),
			H2(
				Class("text-lg font-semibold text-zinc-900 dark:text-zinc-200"),
				g.Text("Ads without embeddings"),
			),
			P(
				Class("text-sm text-zinc-600 dark:text-zinc-400 mt-1"),
				g.Text("Active ads missing a stored vector (up to 25)."),
			),
		),
		embeddingMissingTableHeader(),
		embeddingMissingRows(rows),
	)
}

func embeddingMissingTableHeader() g.Node {
	return Div(
		Class("grid grid-cols-3 gap-2 bg-zinc-50 dark:bg-zinc-900 px-4 py-2 border-b border-zinc-200 dark:border-zinc-700 text-xs"),
		embeddingMissingHeaderCell("ID"),
		embeddingMissingHeaderCell("Title"),
		embeddingMissingHeaderCell("Category"),
	)
}

func embeddingMissingHeaderCell(label string) g.Node {
	return Div(
		Class("font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider"),
		g.Text(label),
	)
}

func embeddingMissingRows(rows []MissingEmbeddingRow) g.Node {
	if len(rows) == 0 {
		return Div(
			Class("px-4 py-6 text-sm text-zinc-500 dark:text-zinc-400"),
			g.Text("All active ads have embeddings."),
		)
	}
	items := make([]g.Node, len(rows))
	for i, row := range rows {
		items[i] = embeddingMissingRow(row)
	}
	return Div(
		ID("embedding-missing-rows"),
		Class("divide-y divide-zinc-200 dark:divide-zinc-700"),
		g.Group(items),
	)
}

func embeddingMissingRow(row MissingEmbeddingRow) g.Node {
	return Div(
		Class("grid grid-cols-3 gap-2 px-4 py-2 text-xs text-zinc-900 dark:text-zinc-200"),
		Div(g.Text(strconv.Itoa(row.AdID))),
		Div(
			Class("truncate"),
			A(
				Href(fmt.Sprintf("/ad/%d", row.AdID)),
				Class("text-blue-600 dark:text-blue-400 hover:underline"),
				g.Text(row.Title),
			),
		),
		Div(g.Text(row.CategoryName)),
	)
}

func embeddingActivitiesSection(d EmbeddingAdminData) g.Node {
	return Div(
		Class("space-y-4"),
		Div(
			Class("flex flex-col sm:flex-row sm:items-end sm:justify-between gap-3"),
			Div(
				H2(
					Class("text-lg font-semibold text-zinc-900 dark:text-zinc-200"),
					g.Text("Search embedding inputs"),
				),
				P(
					Class("text-sm text-zinc-600 dark:text-zinc-400 mt-1"),
					g.Text("Weighted ad activities that feed user and site vectors when the search box is empty."),
				),
			),
			embeddingCategoryFilter(d),
		),
		Div(
			Class("grid grid-cols-1 lg:grid-cols-2 gap-4"),
			embeddingUserActivitiesPanel(d.UserActivities),
			embeddingSiteActivitiesPanel(d.SiteActivities),
		),
	)
}

func embeddingCategoryFilter(d EmbeddingAdminData) g.Node {
	options := make([]g.Node, len(d.Categories))
	for i, cat := range d.Categories {
		options[i] = Option(
			Value(strconv.Itoa(cat.ID)),
			g.If(cat.ID == d.CategoryID, Selected()),
			g.Text(cat.Name),
		)
	}
	return Div(
		Class("flex flex-col gap-1"),
		Label(
			For("embedding-category-filter"),
			Class("text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide"),
			g.Text("Category"),
		),
		Select(
			ID("embedding-category-filter"),
			Name("category"),
			Class("px-3 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md "+
				"bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-200 text-sm"),
			hx.Get("/admin/tab/embeddings"),
			hx.Target("#admin-dashboard-container"),
			hx.Swap("outerHTML"),
			hx.Trigger("change"),
			hx.Include("#embedding-category-filter"),
			g.Group(options),
		),
	)
}

func embeddingUserActivitiesPanel(rows []EmbeddingActivityRow) g.Node {
	return embeddingActivitiesPanel(
		"Your user embedding",
		"Top weighted activity for the logged-in user in this category.",
		rows,
	)
}

func embeddingSiteActivitiesPanel(rows []EmbeddingActivityRow) g.Node {
	return embeddingActivitiesPanel(
		"Site embedding",
		"Top weighted activity across all users in this category.",
		rows,
	)
}

func embeddingActivitiesPanel(title, subtitle string,
	rows []EmbeddingActivityRow) g.Node {
	return Div(
		Class("bg-white dark:bg-zinc-800 rounded-lg shadow overflow-hidden "+
			"border border-zinc-200 dark:border-zinc-700"),
		Div(
			Class("px-4 py-3 border-b border-zinc-200 dark:border-zinc-700"),
			H3(
				Class("text-sm font-semibold text-zinc-900 dark:text-zinc-200"),
				g.Text(title),
			),
			P(
				Class("text-sm text-zinc-600 dark:text-zinc-400 mt-1"),
				g.Text(subtitle),
			),
		),
		embeddingActivitiesTableHeader(),
		embeddingActivityRows(rows),
	)
}

func embeddingActivitiesTableHeader() g.Node {
	return Div(
		Class("grid grid-cols-5 gap-2 bg-zinc-50 dark:bg-zinc-900 px-4 py-2 "+
			"border-b border-zinc-200 dark:border-zinc-700 text-xs"),
		embeddingActivityHeaderCell("Ad"),
		embeddingActivityHeaderCell("Type"),
		embeddingActivityHeaderCell("Weight"),
		embeddingActivityHeaderCell("When"),
		embeddingActivityHeaderCell("ID"),
	)
}

func embeddingActivityHeaderCell(label string) g.Node {
	return Div(
		Class("font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wider"),
		g.Text(label),
	)
}

func embeddingActivityRows(rows []EmbeddingActivityRow) g.Node {
	if len(rows) == 0 {
		return Div(
			Class("px-4 py-6 text-sm text-zinc-500 dark:text-zinc-400"),
			g.Text("No qualifying activity in this category."),
		)
	}
	items := make([]g.Node, len(rows))
	for i, row := range rows {
		items[i] = embeddingActivityRow(row)
	}
	return Div(
		Class("divide-y divide-zinc-200 dark:divide-zinc-700"),
		g.Group(items),
	)
}

func embeddingActivityRow(row EmbeddingActivityRow) g.Node {
	return Div(
		Class("grid grid-cols-5 gap-2 px-4 py-2 text-xs text-zinc-900 dark:text-zinc-200"),
		Div(
			Class("truncate col-span-1"),
			A(
				Href(fmt.Sprintf("/ad/%d", row.AdID)),
				Class("text-blue-600 dark:text-blue-400 hover:underline"),
				g.Text(row.AdTitle),
			),
		),
		Div(g.Text(embeddingActivityTypeLabel(row.ActivityType))),
		Div(g.Text(fmt.Sprintf("%.3f", row.Weight))),
		Div(g.Text(row.Timestamp)),
		Div(g.Text(strconv.Itoa(row.AdID))),
	)
}

func embeddingActivityTypeLabel(activityType string) string {
	switch activityType {
	case "bookmark":
		return "Bookmark"
	case "ad_click":
		return "Ad click"
	case "image_click":
		return "Image click"
	case "ad_created":
		return "Ad created"
	case "recent_ad":
		return "Recent ad"
	default:
		return activityType
	}
}

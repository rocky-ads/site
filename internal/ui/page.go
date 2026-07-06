package ui

import (
	"fmt"
	"strings"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/local"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// seoMetaTags returns SEO meta tags for the homepage
func seoMetaTags() []g.Node {
	return []g.Node{
		Meta(
			Name("description"),
			Content(config.ServerName),
		),
		Meta(
			Name("keywords"),
			Content(
				"classifieds, marketplace, buy and sell, car parts, "+
					"vehicles, jobs, pets, housing, automotive parts, "+
					"bicycles, motorcycles, agricultural equipment",
			),
		),
		Meta(Name("author"), Content(config.ServerName)),
		Meta(Name("robots"), Content("index, follow")),
		Meta(g.Attr("property", "og:title"), Content(config.ServerName)),
		Meta(
			g.Attr("property", "og:description"),
			Content(
				config.ServerName+" - Classified ads without the newspaper. Buy and "+
					"sell vehicles, parts, jobs, pets, housing, and more.",
			),
		),
		Meta(g.Attr("property", "og:type"), Content("website")),
		Meta(
			g.Attr("property", "og:url"),
			Content("https://rockyads.com/"),
		),
		Meta(g.Attr("property", "og:site_name"), Content(config.ServerName)),
		Meta(Name("twitter:card"), Content("summary")),
		Meta(Name("twitter:title"), Content(config.ServerName)),
		Meta(
			Name("twitter:description"),
			Content(config.ServerName),
		),
	}
}

func getUserInitial(userName string) string {
	return strings.ToUpper(string([]rune(userName)[0]))
}

func avatar(userName string) g.Node {
	return Span(
		ID("user-avatar"),
		Class("bg-red-500 text-white rounded-full w-8 h-8 flex items-center justify-center font-semibold text-sm cursor-pointer hover:bg-red-600"),
		hx.Get("/auth/user/menu"),
		hx.Target("body"),
		hx.Swap("beforeend"),
		g.Text(getUserInitial(userName)),
	)
}

func hasUnreadIndicator() g.Node {
	return Div(
		Class("bg-green-500 rounded-full w-3 h-3"),
	)
}

func UnreadIndicatorSwapOOB(hasUnread bool) g.Node {
	return Div(
		hx.SwapOOB("true"),
		ID("message-unread-indicator"),
		g.If(hasUnread, Div(Class("absolute -top-1 -right-3"), hasUnreadIndicator())),
	)
}

func navLoggedIn(userName string, hasUnread bool) g.Node {
	return Div(
		ID("user-avatar-container"),
		Class("relative"),
		avatar(userName),
		UnreadIndicatorSwapOOB(hasUnread),
	)
}

func loginNode() g.Node {
	return A(Href("/login"), Class("text-blue-600 dark:text-blue-400 hover:underline"), g.Text("Login"))
}

func registerNode() g.Node {
	return A(Href("/register"), Class("text-blue-600 dark:text-blue-400 hover:underline"), g.Text("Register"))
}

func navLoggedOut(currentPath string) g.Node {
	switch currentPath {
	case "/login":
		return registerNode()
	case "/register":
		return loginNode()
	case "/register/verify":
		return nil
	default:
		return Div(
			Class("flex items-center space-x-4"),
			loginNode(),
			registerNode(),
		)
	}
}

func indicator() g.Node {
	return Div(
		ID("indicator"),
		Class("htmx-indicator w-4 h-4 border-2 border-blue-600 dark:border-blue-400 border-t-transparent rounded-full animate-spin flex-shrink-0"),
	)
}

// Global listener for unnamed SSE events sunk here.  Assumption is unnamed
// event data contains htmx swap-oob elements to swap into DOM.
func swapOOBmessages() g.Node {
	return Div(
		g.Attr("sse-swap", "message"),
		hx.Swap("none"),
		g.Attr("style", "display: none;"),
	)
}

func RemoveIntroBanner() g.Node {
	return Div(
		ID("intro-banner"),
		hx.SwapOOB("delete"),
	)
}

func introBanner() g.Node {
	return Div(
		ID("intro-banner"),
		Class("relative -mx-4 px-4 h-10 flex items-center "+
			"justify-center bg-amber-100 dark:bg-amber-950/50 "+
			"border-b border-amber-200 dark:border-amber-800 "+
			"text-zinc-700 dark:text-amber-100"),
		A(
			Href("/about"),
			Class("font-medium text-blue-700 dark:text-blue-400 "+
				"hover:underline"),
			hx.Get("/api/intro-banner/dismiss?redirect=/about"),
			hx.Swap("none"),
			g.Text("What am I looking at?"),
		),
		Button(
			Type("button"),
			g.Attr("aria-label", "Close"),
			Class("absolute right-4 top-1/2 -translate-y-1/2 "+
				"w-8 h-8 flex items-center justify-center "+
				"rounded-full hover:bg-amber-200/60 "+
				"dark:hover:bg-amber-900/50 cursor-pointer "+
				"text-xl leading-none"),
			hx.Get("/api/intro-banner/dismiss"),
			hx.Swap("none"),
			g.Text("×"),
		),
	)
}

func navigation(userID int, userName, currentPath string,
	hasUnread bool) g.Node {
	return Nav(
		ID("main-nav"),
		Class("sticky top-0 z-10 "+
			"bg-white/75 dark:bg-zinc-900/75 backdrop-blur-xl "+
			"border-b border-zinc-200 dark:border-zinc-700 "+
			"flex items-center justify-between mb-4 py-4 -mx-4 px-4"),
		A(
			Href("/"),
			Class("flex items-center gap-2 shrink-0 whitespace-nowrap"),
			Img(
				Src("/images/rock.svg"),
				Alt(""),
				Class("w-9 h-9 flex-shrink-0"),
			),
			Span(Class("text-xl font-bold"), g.Text(config.ServerName)),
		),
		Div(
			Class("flex items-center gap-3 shrink-0"),
			indicator(),
			g.Iff(local.IsLoggedIn(userID), func() g.Node { return navLoggedIn(userName, hasUnread) }),
			g.Iff(!local.IsLoggedIn(userID), func() g.Node { return navLoggedOut(currentPath) }),
		),
	)
}

func Page(userID int, hasUnread bool, userName, title, currentPath,
	csrfToken string, showIntroBanner bool, body []g.Node) g.Node {
	var headNodes []g.Node

	// SEO meta tags for homepage
	if currentPath == "/" {
		headNodes = append(headNodes, g.Group(seoMetaTags()))
	}

	// Favicon
	headNodes = append(headNodes,
		Link(
			Rel("icon"),
			Type("image/svg+xml"),
			Href("/images/rock.svg"),
		),
	)

	// Stylesheets
	headNodes = append(headNodes,
		Link(
			Rel("stylesheet"),
			Href("/css/output.css"),
		),
		Script(
			Type("text/javascript"),
			Src(config.HTMXURL),
			Defer(),
		),
		Script(
			Type("text/javascript"),
			Src(config.HTMXSSEURL),
			Defer(),
		),
		Script(
			Type("text/javascript"),
			Src("/js/timezone.js"),
			Defer(),
		),
		Script(
			Type("text/javascript"),
			Src("/js/theme.js"),
			Defer(),
		),
	)

	// Properly escape the CSRF token for JSON
	headersJSON := fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)

	contentClass := "w-full md:max-w-3xl md:mx-auto pb-8 px-6 " +
		"text-zinc-900 dark:text-zinc-200 bg-white dark:bg-zinc-900"

	bodyNodes := []g.Node{
		Class("min-h-screen bg-white dark:bg-zinc-900"),
		Div(
			ID("page-content"),
			g.If(local.IsLoggedIn(userID), hx.Ext("sse")),
			g.If(local.IsLoggedIn(userID), g.Attr("sse-connect", "/auth/sse")),
			g.If(local.IsLoggedIn(userID), g.Attr("sse-close", "close")),
			g.If(local.IsLoggedIn(userID), swapOOBmessages()),
			Div(
				Class(contentClass),
				hx.Headers(headersJSON),
				hx.Indicator("#indicator"),
				g.If(showIntroBanner, introBanner()),
				navigation(userID, userName, currentPath, hasUnread),
				g.Group(body),
			),
		),
	}

	docTitle := config.ServerName
	if title != "" && title != config.ServerName {
		docTitle = config.ServerName + " - " + title
	}

	return components.HTML5(components.HTML5Props{
		Title:    docTitle,
		Language: "en",
		Head:     headNodes,
		Body:     bodyNodes,
	})
}

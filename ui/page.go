package ui

import (
	"fmt"
	"strings"

	"github.com/rocky-ads/site/config"
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
				"Rocky Ads - Classified ads without the newspaper. Buy and "+
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

func navLoggedIn(userName string) g.Node {
	return Div(
		Class("relative"),
		Span(
			ID("user-avatar"),
			Class("bg-red-500 text-white rounded-full w-8 h-8 flex items-center justify-center font-semibold text-sm cursor-pointer hover:bg-red-600"),
			hx.Get("/auth/user-menu"),
			hx.Swap("none"),
			g.Text(getUserInitial(userName)),
		),
	)
}

func loginNode() g.Node {
	return A(Href("/login"), Class("text-blue-500 hover:underline"), g.Text("Login"))
}

func registerNode() g.Node {
	return A(Href("/register"), Class("text-blue-500 hover:underline"), g.Text("Register"))
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
		Class("htmx-indicator flex items-center gap-2 text-blue-600"),
		Div(
			Class("w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"),
		),
		g.Text("Working..."),
	)
}

func navigation(userID int, userName, currentPath string) g.Node {
	return Nav(
		Class("sticky top-0 z-10 bg-white/75 dark:bg-gray-900/75 backdrop-blur-xl border-b border-gray-200 dark:border-gray-700 flex items-center justify-between mb-8 py-4 -mx-4 px-4"),
		Div(
			Class("flex items-center gap-4"),
			A(Href("/"), Class("text-xl font-bold"), g.Text(config.ServerName)),
		),
		indicator(),
		g.Iff(userID != 0, func() g.Node { return navLoggedIn(userName) }),
		g.Iff(userID == 0, func() g.Node { return navLoggedOut(currentPath) }),
	)
}

func Page(userID int, userName, title, currentPath, csrfToken string, body []g.Node) g.Node {
	var headNodes []g.Node

	// SEO meta tags for homepage
	if currentPath == "/" {
		headNodes = append(headNodes, g.Group(seoMetaTags()))
	}

	// Favicons
	headNodes = append(headNodes,
		Link(
			Rel("icon"),
			Type("image/png"),
			Href("/images/favicon-32x32.png"),
			g.Attr("sizes", "32x32"),
		),
		Link(
			Rel("icon"),
			Type("image/png"),
			Href("/images/favicon-16x16.png"),
			g.Attr("sizes", "16x16"),
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

	return components.HTML5(components.HTML5Props{
		Title:    config.ServerName,
		Language: "en",
		Head:     headNodes,
		Body: []g.Node{
			Class("min-h-screen bg-white dark:bg-gray-900"),
			Div(
				Class("w-full md:max-w-3xl md:mx-auto pb-8 px-4 text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-900"),
				hx.Headers(headersJSON),
				hx.Indicator("#indicator"),
				navigation(userID, userName, currentPath),
				g.Group(body),
			),
			g.Group(modalPlaceholder()),
		},
	})
}

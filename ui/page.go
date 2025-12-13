package ui

import (
	"fmt"

	"github.com/rocky-ads/site/config"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	"maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// Page returns a complete HTML page for MPA (Multi-Page Application)
func Page(title, currentPath, csrfToken string, body []g.Node) g.Node {
	var headNodes []g.Node

	// SEO meta tags for homepage
	if currentPath == "/" {
		headNodes = append(headNodes, g.Group(seoMetaTags(title, currentPath)))
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
		Title:    title,
		Language: "en",
		Head:     headNodes,
		Body: []g.Node{
			Class("min-h-screen bg-white dark:bg-gray-900"),
			Div(
				Class("w-full md:max-w-4xl md:mx-auto py-8 px-4 text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-900"),
				hx.Headers(headersJSON),
				navigation(),
				g.Group(body),
			),
		},
	})
}

// seoMetaTags returns SEO meta tags for the homepage
func seoMetaTags(title, currentPath string) []g.Node {
	return []g.Node{
		Meta(
			Name("description"),
			Content(
				"Rocky Ads - Classified ads without the newspaper",
			),
		),
		Meta(
			Name("keywords"),
			Content(
				"classifieds, marketplace, buy and sell, car parts, "+
					"vehicles, jobs, pets, housing, automotive parts, "+
					"bicycles, motorcycles, agricultural equipment",
			),
		),
		Meta(Name("author"), Content("Rocky Ads")),
		Meta(Name("robots"), Content("index, follow")),
		Meta(g.Attr("property", "og:title"), Content(title)),
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
			Content("https://rockyads.com"+currentPath),
		),
		Meta(g.Attr("property", "og:site_name"), Content("Rocky Ads")),
		Meta(Name("twitter:card"), Content("summary")),
		Meta(Name("twitter:title"), Content(title)),
		Meta(
			Name("twitter:description"),
			Content(
				"Rocky Ads - Classified ads without the newspaper. Buy and "+
					"sell vehicles, parts, jobs, pets, housing, and more.",
			),
		),
	}
}

// navigation returns the navigation component
func navigation() g.Node {
	return Div(
		Class("flex items-center justify-between mb-8 pb-4 border-b border-gray-200 dark:border-gray-700"),
		Div(
			Class("text-2xl font-bold text-gray-900 dark:text-gray-100"),
			g.Text("Rocky Ads"),
		),
	)
}

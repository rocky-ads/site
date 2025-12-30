package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func WelcomePage(rockCount int) []g.Node {
	return []g.Node{
		pageTitle("Welcome"),
		Div(
			Class("space-y-8"),
			Div(
				Class("bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6 space-y-4 text-center"),
				Div(
					Class("flex justify-center mb-4"),
					Img(
						Src("/images/rocks.png"),
						Alt("Three stacked rocks"),
						Class("w-24"),
					),
				),
				P(
					Class("text-zinc-900 dark:text-zinc-200 text-lg"),
					g.Text("Welcome to Rocky Ads!"),
				),
				P(
					Class("text-zinc-900 dark:text-zinc-200 text-lg"),
					g.Textf("You've been given %d rocks to help maintain quality on our platform.", rockCount),
				),
			),
			Div(
				Class("space-y-4"),
				H2(
					Class("text-2xl font-semibold"),
					g.Text("How Rocks Work"),
				),
				Ul(
					Class("space-y-2 list-disc list-inside"),
					Li(g.Text("Throw rocks at ads that violate our policies or have issues")),
					Li(g.Text("Each rock creates a conversation with the ad owner")),
					Li(g.Text("Work together to resolve the dispute")),
					Li(g.Text("Once resolved, the seller can return your rock")),
					Li(g.Text("Rocks are limited - use them wisely!")),
				),
			),
		),
	}
}

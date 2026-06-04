package ui

import (
	"github.com/rocky-ads/site/internal/config"
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func WelcomePage(eggCount int) []g.Node {
	return []g.Node{
		Div(
			Class("space-y-8"),
			// Welcome card with hero section
			Div(
				Class("bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/30 dark:to-indigo-900/30 border border-blue-200 dark:border-blue-800 rounded-xl p-8 md:p-12 shadow-lg"),
				Div(
					Class("flex flex-col items-center space-y-6 text-center"),
					Div(
						Class("flex justify-center mb-2"),
						Img(
							Src("/images/three-eggs.png"),
							Alt("Three stacked eggs"),
							Class("w-full max-w-[280px] md:max-w-[350px] rounded-xl shadow-md"),
						),
					),
					Div(
						Class("space-y-4"),
						H1(
							Class("text-3xl md:text-4xl font-bold text-zinc-900 dark:text-zinc-100"),
							g.Textf("Welcome to %s!", config.ServerName),
						),
						P(
							Class("text-lg md:text-xl text-zinc-700 dark:text-zinc-300 font-medium"),
							g.Textf("You've been given %d eggs to help maintain quality on our platform", eggCount),
						),
						A(
							Href("/"),
							Class("inline-block px-6 py-3 bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 text-white font-semibold rounded-lg shadow-md hover:shadow-lg transition-all duration-200"),
							g.Text("Go To Ads"),
						),
					),
				),
			),
			// How it works section
			Div(
				Class("bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-xl p-6 md:p-8 shadow-sm"),
				Div(
					Class("space-y-6"),
					H2(
						Class("text-2xl md:text-3xl font-bold text-zinc-900 dark:text-zinc-100 mb-2"),
						g.Text("How Eggs Work"),
					),
					Ul(
						Class("space-y-4 list-none"),
						Li(
							Class("flex items-start gap-3"),
							Span(Class("flex-shrink-0 w-6 h-6 rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 flex items-center justify-center font-semibold text-sm mt-0.5"), g.Text("1")),
							Span(Class("text-zinc-700 dark:text-zinc-300 text-base leading-relaxed"), g.Text("Throw eggs at ads that violate our policies or have issues")),
						),
						Li(
							Class("flex items-start gap-3"),
							Span(Class("flex-shrink-0 w-6 h-6 rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 flex items-center justify-center font-semibold text-sm mt-0.5"), g.Text("2")),
							Span(Class("text-zinc-700 dark:text-zinc-300 text-base leading-relaxed"), g.Text("Each egg creates a conversation with the ad owner")),
						),
						Li(
							Class("flex items-start gap-3"),
							Span(Class("flex-shrink-0 w-6 h-6 rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 flex items-center justify-center font-semibold text-sm mt-0.5"), g.Text("3")),
							Span(Class("text-zinc-700 dark:text-zinc-300 text-base leading-relaxed"), g.Text("Work together to resolve the dispute")),
						),
						Li(
							Class("flex items-start gap-3"),
							Span(Class("flex-shrink-0 w-6 h-6 rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 flex items-center justify-center font-semibold text-sm mt-0.5"), g.Text("4")),
							Span(Class("text-zinc-700 dark:text-zinc-300 text-base leading-relaxed"), g.Text("Once resolved, the seller can return your egg")),
						),
						Li(
							Class("flex items-start gap-3"),
							Span(Class("flex-shrink-0 w-6 h-6 rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 flex items-center justify-center font-semibold text-sm mt-0.5"), g.Text("5")),
							Span(Class("text-zinc-700 dark:text-zinc-300 text-base leading-relaxed font-medium"), g.Text("Eggs are limited - use them wisely!")),
						),
					),
				),
			),
		),
	}
}

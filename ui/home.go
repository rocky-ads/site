package ui

import (
	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// HomePageContent returns the content for the homepage
func HomePageContent() []g.Node {
	return []g.Node{
		H1(g.Text("Welcome to Rocky Ads")),
		P(g.Text("Classified ads without the newspaper.")),
	}
}

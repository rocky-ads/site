package ui

import (
	"fmt"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// renderEggIconsByOrdinal renders clickable egg icons using ordinal positions
func renderEggIconsByOrdinal(eggCount int, urlPattern string, id int) g.Node {
	if eggCount <= 0 {
		return g.Raw("")
	}

	var icons []g.Node
	for i := range eggCount {
		icons = append(icons,
			Div(
				Class("egg-icon-container flex items-center gap-0.5 cursor-pointer hover:opacity-80 transition-opacity"),
				g.Attr("onclick", "event.stopPropagation();"),
				hx.Get(fmt.Sprintf(urlPattern, id, i)),
				hx.Target("body"),
				hx.Swap("beforeend"),
				Img(
					Src("/images/broken-egg.svg"),
					Alt("Egg thrown"),
					Class("w-5 h-5 flex-shrink-0"),
					Title("Egg thrown - click to view dispute assessment"),
				),
			),
		)
	}

	return Div(
		Class("flex items-center gap-0.5"),
		g.Group(icons),
	)
}

// EggIcons renders clickable egg icons for an ad
// Each icon links to an egg by ordinal position
// Only shows eggs bound to the ad (when inquirer throws)
func EggIcons(adID int, eggCount int) g.Node {
	return renderEggIconsByOrdinal(eggCount, "/auth/ad/%d/egg/%d", adID)
}

// UserEggIcons renders clickable egg icons for a user
// Each icon links to an egg by ordinal position
// Shows eggs bound to the user (when owner throws)
func UserEggIcons(userID int, eggCount int) g.Node {
	return renderEggIconsByOrdinal(eggCount, "/auth/user/%d/egg/%d", userID)
}

// EggThrowLink renders a link to throw/unthrow an egg in the conversation modal
// hasThrownEgg: whether the current user has thrown an egg at this conversation
// canThrow: whether the current user can throw an egg (is participant and has < 3 eggs)
func EggThrowLink(conversationID int, hasThrownEgg, canThrow bool,
	csrfToken string) g.Node {
	// Only show link if user has thrown an egg (to remove it) OR user can throw an egg (and hasn't thrown one)
	if !hasThrownEgg && !canThrow {
		return g.Raw("")
	}

	var attrs []g.Node
	var text string
	var class string

	if hasThrownEgg {
		// Unthrow link - user has thrown an egg, show remove option
		text = "Remove Egg"
		class = "text-red-600 dark:text-red-400 hover:underline text-sm"
		attrs = []g.Node{
			hx.Delete(fmt.Sprintf("/auth/conversation/%d/egg/unthrow", conversationID)),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Target("body"),
			hx.Swap("outerHTML"),
		}
	} else if canThrow {
		// Throw link - user can throw and hasn't thrown one yet
		text = "Throw Egg"
		class = "text-orange-600 dark:text-orange-400 hover:underline text-sm"
		attrs = []g.Node{
			hx.Post(fmt.Sprintf("/auth/conversation/%d/egg/throw", conversationID)),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Target("body"),
			hx.Swap("outerHTML"),
		}
	} else {
		// Shouldn't reach here due to check above, but just in case
		return g.Raw("")
	}

	return A(
		g.Group(attrs),
		Href("#"),
		Class(class),
		g.Text(text),
	)
}

// EggCountBadge shows user's remaining eggs
func EggCountBadge(eggCount int) g.Node {
	if eggCount <= 0 {
		return g.Raw("")
	}

	return Span(
		Class("px-2 py-1 bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 rounded-full text-xs font-medium"),
		g.Text(fmt.Sprintf("%d eggs", eggCount)),
	)
}

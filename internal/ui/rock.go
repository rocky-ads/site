package ui

import (
	"fmt"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// renderRockIconsByOrdinal renders clickable rock icons using ordinal positions
func renderRockIconsByOrdinal(rockCount int, urlPattern string, id int) g.Node {
	if rockCount <= 0 {
		return g.Raw("")
	}

	var icons []g.Node
	for i := range rockCount {
		icons = append(icons,
			Div(
				Class("rock-icon-container flex items-center gap-0.5 cursor-pointer hover:opacity-80 transition-opacity"),
				g.Attr("onclick", "event.stopPropagation();"),
				hx.Get(fmt.Sprintf(urlPattern, id, i)),
				hx.Target("body"),
				hx.Swap("beforeend"),
				Img(
					Src("/images/rock.svg"),
					Alt("Rock thrown"),
					Class("w-5 h-5 flex-shrink-0"),
					Title("Rock thrown - click to view dispute assessment"),
				),
			),
		)
	}

	return Div(
		Class("flex items-center gap-0.5"),
		g.Group(icons),
	)
}

// RockIcons renders clickable rock icons for an ad
// Each icon links to a rock by ordinal position
// Only shows rocks bound to the ad (when inquirer throws)
func RockIcons(adID int, rockCount int) g.Node {
	return renderRockIconsByOrdinal(rockCount, "/auth/ad/%d/rock/%d", adID)
}

// UserRockIcons renders clickable rock icons for a user
// Each icon links to a rock by ordinal position
// Shows rocks bound to the user (when owner throws)
func UserRockIcons(userID int, rockCount int) g.Node {
	return renderRockIconsByOrdinal(rockCount, "/auth/user/%d/rock/%d", userID)
}

// RockThrowLink renders a link to throw/unthrow a rock in the conversation modal
// hasThrownRock: whether the current user has thrown a rock at this conversation
// canThrow: whether the current user can throw a rock (is participant and has < 3 rocks)
func RockThrowLink(conversationID int, hasThrownRock, canThrow bool,
	csrfToken string) g.Node {
	// Only show link if user has thrown a rock (to remove it) OR user can throw a rock (and hasn't thrown one)
	if !hasThrownRock && !canThrow {
		return g.Raw("")
	}

	var attrs []g.Node
	var text string
	var class string

	if hasThrownRock {
		// Unthrow link - user has thrown a rock, show remove option
		text = "Remove Rock"
		class = "text-red-600 dark:text-red-400 hover:underline text-sm"
		attrs = []g.Node{
			hx.Delete(fmt.Sprintf("/auth/conversation/%d/rock/unthrow", conversationID)),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Target("body"),
			hx.Swap("outerHTML"),
		}
	} else if canThrow {
		// Throw link - user can throw and hasn't thrown one yet
		text = "Throw Rock"
		class = "text-orange-600 dark:text-orange-400 hover:underline text-sm"
		attrs = []g.Node{
			hx.Post(fmt.Sprintf("/auth/conversation/%d/rock/throw", conversationID)),
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

// RockCountBadge shows user's remaining rocks
func RockCountBadge(rockCount int) g.Node {
	if rockCount <= 0 {
		return g.Raw("")
	}

	return Span(
		Class("px-2 py-1 bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200 rounded-full text-xs font-medium"),
		g.Text(fmt.Sprintf("%d rocks", rockCount)),
	)
}

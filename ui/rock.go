package ui

import (
	"fmt"

	"github.com/rocky-ads/site/rock"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

// RockIcons renders clickable rock icons for an ad
// Each icon links directly to its conversation
func RockIcons(adID int, count int, currentUserID int) g.Node {
	if count <= 0 {
		return g.Raw("")
	}

	// Get all conversation IDs for this ad
	conversationIDs, err := rock.GetPublicConversationsForAd(adID)
	if err != nil || len(conversationIDs) == 0 {
		return g.Raw("")
	}

	var icons []g.Node
	for _, conversationID := range conversationIDs {
		icons = append(icons,
			Div(
				Class("rock-icon-container flex items-center gap-0.5 cursor-pointer hover:opacity-80 transition-opacity"),
				g.Attr("onclick", "event.stopPropagation();"),
				hx.Get(fmt.Sprintf("/auth/conversation/%d", conversationID)),
				hx.Target("body"),
				hx.Swap("beforeend"),
				Img(
					Src("/images/broken-egg.svg"),
					Alt("Rock thrown"),
					Class("w-5 h-5 flex-shrink-0"),
					Title("Rock thrown - click to view conversation"),
				),
			),
		)
	}

	return Div(
		Class("flex items-center gap-0.5"),
		g.Group(icons),
	)
}

// RockThrowButton renders a button to throw/unthrow a rock in the conversation modal
// hasThrownRock: whether the current user has thrown a rock at this conversation
// canThrow: whether the current user can throw a rock (is participant and has < 3 rocks)
// Note: Multiple users can throw rocks at the same conversation, so the button shows "Throw Rock"
//
//	even if someone else has already thrown a rock (as long as current user hasn't)
func RockThrowButton(conversationID int, hasThrownRock, canThrow bool, csrfToken string) g.Node {
	// Only show button if user has thrown a rock (to remove it) OR user can throw a rock (and hasn't thrown one)
	if !hasThrownRock && !canThrow {
		return g.Raw("")
	}

	var attrs []g.Node
	var text string
	var class string

	if hasThrownRock {
		// Unthrow button - user has thrown a rock, show remove option
		text = "Remove Rock"
		class = "px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors text-sm"
		attrs = []g.Node{
			hx.Delete(fmt.Sprintf("/auth/conversation/%d/rock/unthrow", conversationID)),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Target("body"),
			hx.Swap("outerHTML"),
		}
	} else if canThrow {
		// Throw button - user can throw and hasn't thrown one yet
		text = "Throw Rock"
		class = "px-4 py-2 bg-orange-500 text-white rounded-md hover:bg-orange-600 transition-colors text-sm"
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

	return Button(
		g.Group(attrs),
		Type("button"),
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

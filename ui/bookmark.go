package ui

import (
	"fmt"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func Bookmark(adID int, bookmarked bool, csrfToken string) g.Node {
	return Button(
		Type("button"),
		g.If(bookmarked, hx.Delete(fmt.Sprintf("/auth/bookmark/%d", adID))),
		g.If(!bookmarked, hx.Post(fmt.Sprintf("/auth/bookmark/%d", adID))),
		hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		g.Attr("onclick", "event.stopPropagation()"),
		g.If(bookmarked, Img(
			Class("w-6 h-6"),
			Src("/images/bookmark-true.svg"),
			Alt("Bookmark"),
		)),
		g.If(!bookmarked, Img(
			Class("w-6 h-6 dark:invert"),
			Src("/images/bookmark-false.svg"),
			Alt("Bookmark"),
		)),
	)
}

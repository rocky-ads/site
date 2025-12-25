package ui

import (
	"fmt"
	"strconv"
	"time"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func Bookmark(adID int, bookmarked bool, csrfToken string) g.Node {
	return Button(
		Type("button"),
		Class("flex-shrink-0"),
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

func AdGridNode(adID int, title string) g.Node {
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class("block"),
		ID(fmt.Sprintf("ad-%d", adID)),
		g.Text(title),
	)
}

func AdListNode(userID, adID, price int, title, location string, createdAt time.Time, active, bookmarked bool, csrfToken string) g.Node {
	class := "flex flex-wrap items-center justify-between py-2 px-3 cursor-pointer"
	if active {
		class += " hover:bg-gray-50 dark:hover:bg-gray-800"
	} else {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		ID(fmt.Sprintf("ad-%d", adID)),
		Div(
			Class("flex items-center gap-2 text-blue-600 hover:text-blue-800 min-w-0"),
			g.If(userID != 0, Bookmark(adID, bookmarked, csrfToken)),
			Span(Class("min-w-0"), g.Text(title)),
		),
		Div(
			Class("flex items-center gap-2 ml-auto"),
			Div(
				Class("flex items-center gap-2 text-xs text-gray-500"),
				ageNode(createdAt),
				locationNode(location),
			),
			priceNode(price),
		),
	)
}

func AdTreeNode(adID int, title string) g.Node {
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class("block"),
		ID(fmt.Sprintf("ad-%d", adID)),
		g.Text(title),
	)
}

// format ad age as Xm, XhYm, Xd, Xmo, or Xy Xmo
func formatAdAge(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}

	days := int(d.Hours() / 24)
	if days <= 31 {
		return fmt.Sprintf("%dd", days)
	}

	// Calculate months and years
	now := time.Now()
	years := now.Year() - t.Year()
	months := int(now.Month()) - int(t.Month())

	// Adjust for day of month
	if now.Day() < t.Day() {
		months--
	}

	// Adjust years if months went negative
	if months < 0 {
		years--
		months += 12
	}

	if years > 0 {
		if months > 0 {
			return fmt.Sprintf("%dy %dmo", years, months)
		}
		return fmt.Sprintf("%dy", years)
	}

	return fmt.Sprintf("%dmo", months)
}

func newBadge() g.Node {
	return Span(
		Class("px-2 py-0.5 rounded-full border border-orange-500 text-orange-500 text-xs font-medium"),
		g.Text("New!"),
	)
}

func ageNode(createdAt time.Time) g.Node {
	return Div(
		Class("flex items-center gap-2"),
		g.If(time.Since(createdAt) < 4*time.Hour, newBadge()),
		g.Text(formatAdAge(createdAt)),
	)
}

func locationNode(location string) g.Node {
	return Span(Class("text-xs text-gray-500"), g.Text(location))
}

func priceNode(price int) g.Node {
	return Span(Class("text-green-600 font-semibold"), g.Text(fmt.Sprintf("$%.0f", float64(price)/100)))
}

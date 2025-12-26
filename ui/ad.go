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

func gridNoImage() g.Node {
	return Div(
		Class("rounded-md w-full h-48 flex items-center justify-center border-2 border-dotted border-gray-300 dark:border-gray-600"),
		g.Text("No Image"),
	)
}

func gridImageCount(current, count int) g.Node {
	return Span(
		Class("absolute bottom-2 right-2 bg-black/50 text-white text-xs px-2 py-1 rounded-full"),
		g.Text(fmt.Sprintf("%d/%d", current, count)),
	)
}

func gridImageNav(adID, current, count int) g.Node {

	prevIdx := (current-2+count)%count + 1
	nextIdx := current%count + 1

	return g.Group([]g.Node{
		// Left button
		Button(
			Type("button"),
			Class("absolute left-2 top-1/2 transform -translate-y-1/2 bg-white/50 rounded-full w-10 h-10 flex items-center justify-center shadow-lg hover:bg-white/60 focus:outline-none cursor-pointer z-20 opacity-100 md:opacity-0 md:group-hover:opacity-100 md:transition-opacity"),
			hx.Get(fmt.Sprintf("/api/grid-image/%d/%d/%d", adID, prevIdx, count)),
			hx.Target(fmt.Sprintf("#grid-image-%d", adID)),
			hx.Swap("outerHTML"),
			g.Attr("onclick", "event.stopPropagation()"),
			Img(
				Class("w-6 h-6"),
				Src("/images/left.svg"),
			),
		),
		// Right button
		Button(
			Type("button"),
			Class("absolute right-2 top-1/2 transform -translate-y-1/2 bg-white/50 rounded-full w-10 h-10 flex items-center justify-center shadow-lg hover:bg-white/60 focus:outline-none cursor-pointer z-20 opacity-100 md:opacity-0 md:group-hover:opacity-100 md:transition-opacity"),
			hx.Get(fmt.Sprintf("/api/grid-image/%d/%d/%d", adID, nextIdx, count)),
			hx.Target(fmt.Sprintf("#grid-image-%d", adID)),
			hx.Swap("outerHTML"),
			g.Attr("onclick", "event.stopPropagation()"),
			Img(
				Class("w-6 h-6"),
				Src("/images/right.svg"),
			),
		),
	})
}

func GridImage(adID, count, current int) g.Node {
	return Div(
		ID(fmt.Sprintf("grid-image-%d", adID)),
		Class("relative group"),
		Img(
			Class("rounded-md w-full h-48 object-cover"),
			Src(fmt.Sprintf("/image/%d/%d/480w", adID, current)),
			g.Attr("loading", "lazy"),
		),
		g.If(count > 1, gridImageCount(current, count)),
		g.If(count > 1, gridImageNav(adID, current, count)),
	)
}

func gridImageNode(adID, count, current int, title string) g.Node {
	if count == 0 {
		return gridNoImage()
	}
	return GridImage(adID, count, current)
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

func AdGridNode(userID, adID, price, imageCount int, title, location string, createdAt time.Time, active, bookmarked bool, csrfToken string) g.Node {
	priceStr := fmt.Sprintf("$%.0f", float64(price)/100)
	class := "flex flex-col cursor-pointer py-1"
	if !active {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		gridImageNode(adID, imageCount, 1, title),
		Span(Class("text-green-600 font-semibold pt-2"), g.Text(priceStr)),
		Span(Class("min-w-0"), g.Text(title)),
		Div(
			Class("flex items-center gap-2"),
			Div(
				Class("flex items-center gap-2 text-xs text-gray-500"),
				ageNode(createdAt),
				g.Text(location),
			),
			g.If(userID != 0, Bookmark(adID, bookmarked, csrfToken)),
		),
	)
}

func AdListNode(userID, adID, price int, title, location string, createdAt time.Time, active, bookmarked bool, csrfToken string) g.Node {
	priceStr := fmt.Sprintf("$%.0f", float64(price)/100)
	class := "flex flex-wrap items-center justify-between py-2 px-3 cursor-pointer"
	if active {
		class += " hover:bg-gray-50 dark:hover:bg-gray-800"
	} else {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}
	return A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		Div(
			Class("flex items-center gap-2 min-w-0"),
			g.If(userID != 0, Bookmark(adID, bookmarked, csrfToken)),
			Span(Class("min-w-0"), g.Text(title)),
		),
		Div(
			Class("flex items-center gap-2 ml-auto"),
			Div(
				Class("flex items-center gap-2 text-xs text-gray-500"),
				ageNode(createdAt),
				g.Text(location),
			),
			Span(Class("text-green-600 font-semibold"), g.Text(priceStr)),
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

package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rocky-ads/site/config"
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

func gridImageNode(adID, count, current int, title string) g.Node {
	if count == 0 {
		return noImage("h-48")
	}
	return ImageNode(adID, count, current, "480w", "h-48", false)
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

func AdGridNode(userID, adID, price, imageCount, nextPage int, title, location, csrfToken string, createdAt time.Time, active, bookmarked, isLast bool) g.Node {
	priceStr := fmt.Sprintf("$%.0f", float64(price)/100)

	class := "flex flex-col cursor-pointer gap-1 py-3"
	if !active {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}

	node := A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		gridImageNode(adID, imageCount, 1, title),
		Div(
			Class("flex items-center justify-between pt-1"),
			Span(Class("text-green-600 font-semibold"), g.Text(priceStr)),
			Div(
				Class("flex gap-2 text-xs text-gray-500"),
				ageNode(createdAt),
				g.Text(location),
			),
		),
		Span(Class("min-w-0"), g.Text(title)),
	)

	if isLast {
		return g.Group([]g.Node{
			node,
			Div(
				hx.Get(fmt.Sprintf("/search/?page=%d", nextPage)),
				hx.Trigger("revealed"),
				hx.Swap("afterend"),
				hx.Include("form"),
			),
		})
	}

	return node
}

func AdListNode(userID, adID, price int, title, location string, createdAt time.Time, active, bookmarked bool, csrfToken string, isLast bool, nextPage int) g.Node {
	priceStr := fmt.Sprintf("$%.0f", float64(price)/100)

	class := "flex flex-wrap items-center justify-between py-2 px-3 cursor-pointer"
	if active {
		class += " hover:bg-gray-50 dark:hover:bg-gray-800"
	} else {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}

	node := A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		Div(
			Class("flex items-center gap-2 min-w-0"),
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

	if isLast {
		return g.Group([]g.Node{
			node,
			Div(
				hx.Get(fmt.Sprintf("/search/?page=%d", nextPage)),
				hx.Trigger("revealed"),
				hx.Swap("afterend"),
				hx.Include("form"),
			),
		})
	}

	return node
}

func deletedWatermark() g.Node {
	return Div(
		Class("absolute top-0 left-0 right-0 bottom-0 flex items-center justify-center pointer-events-none z-50"),
		Div(
			Class("font-bold text-8xl text-red-500 transform rotate-[-45deg]"),
			g.Text("DELETED"),
		),
	)
}

func Ad(adID, userID, imageCount, price int, title, location, description string,
	createdAt time.Time, bookmarked, active bool, csrfToken string) []g.Node {

	priceStr := fmt.Sprintf("$%.0f", float64(price)/100)

	return []g.Node{
		Div(
			Class("flex flex-col relative rounded-lg shadow-xl/50 my-4 mx-2 col-span-full overflow-hidden"),
			g.If(imageCount > 0, ImageNode(adID, imageCount, 1, "1200w", "h-96 md:h-[600px] lg:h-[600px]", true)),
			g.If(imageCount == 0, noImage("h-96 md:h-[600px] lg:h-[600px]")),
			g.If(!active, deletedWatermark()),
			Div(
				Class("p-4 flex flex-col"),
				Div(
					Class("flex items-center gap-2 min-w-0"),
					g.If(userID != 0, Bookmark(adID, bookmarked, csrfToken)),
					Span(Class("min-w-0"), g.Text(title)),
				),
				Div(
					Class("flex items-center gap-2"),
					Div(
						Class("flex items-center gap-2 text-xs text-gray-500"),
						ageNode(createdAt),
						g.Text(location),
					),
					Span(Class("text-green-600 font-semibold"), g.Text(priceStr)),
				),
				Div(Class("text-base mt-2 whitespace-pre-wrap"), g.Text(description)),
			),
		),
	}
}

func AdDeleted() []g.Node {
	return []g.Node{
		Div(
			Class("text-center py-16"),
			Div(
				Class("mb-6 flex justify-center"),
				Img(
					Src("/images/trashcan.svg"),
					Alt("Deleted"),
					Class("w-24 h-24"),
				),
			),
			H2(
				Class("text-3xl font-bold text-gray-900 dark:text-gray-100 mb-4"),
				g.Text("Ad Deleted"),
			),
			P(
				Class("text-lg text-gray-600 dark:text-gray-400 mb-8"),
				g.Text("This ad has been deleted by the owner and is no longer available."),
			),
			standardButton(buttonProps{
				Text: "Back to Home",
				Href: "/",
			}),
		),
	}
}

func categoryNode(categoryName string) g.Node {
	return Div(
		Class("text-lg text-gray-600 dark:text-gray-400 italic"),
		g.Text(categoryName),
	)
}

func titleInput() g.Node {
	return Div(
		label("Title"),
		inputText("title", "", true,
			MaxLength("35"),
			Pattern("[\\x20-\\x7E]+"),
			//Title("Title must be 1-35 characters, printable ASCII characters only"),
			g.Attr("oninput", "this.checkValidity()"),
		),
	)
}

func descriptionInput() g.Node {
	return Div(
		label("Description"),
		textArea("description", "", true,
			MaxLength(fmt.Sprintf("%d", config.MaxAdDescriptionLength)),
			Rows("4"),
			Pattern("[\\x20-\\x7E\\n\\r]+"),
			Title("Description must contain printable ASCII characters only"),
			g.Attr("oninput", "this.checkValidity()"),
		),
	)
}

func imagesInput() g.Node {
	return Div(
		label("Images"),
		inputText("images", "", true),
	)
}

func newAdForm(fields []g.Node) g.Node {
	return Form(
		Class("space-y-8 mt-8"),
		hx.Post("/api/ad/new"),
		hx.Swap("none"),
		titleInput(),
		g.Group(fields),
		imagesInput(),
		descriptionInput(),
		standardButton(buttonProps{
			Type: "submit",
			Text: "Submit",
		}),
	)
}

func NewAd(categoryName string, fields []g.Node) []g.Node {
	return []g.Node{
		pageTitle("Create New Ad"),
		categoryNode(categoryName),
		newAdForm(fields),
	}
}

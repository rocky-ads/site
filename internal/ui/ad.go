package ui

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/local"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func BookmarkButton(adID int, bookmarked bool, csrfToken string) g.Node {
	var imageNode g.Node
	if bookmarked {
		imageNode = Img(
			Class("w-6 h-6"),
			Src("/images/bookmark-true.svg"),
			Alt("Bookmark"),
		)
	} else {
		imageNode = Img(
			Class("w-6 h-6 dark:invert dark:opacity-80"),
			Src("/images/bookmark-false.svg"),
			Alt("Bookmark"),
		)
	}

	return iconButton(buttonProps{
		Children: []g.Node{imageNode},
		Attrs: []g.Node{
			g.If(bookmarked, hx.Delete(fmt.Sprintf("/auth/bookmark/%d", adID))),
			g.If(!bookmarked, hx.Post(fmt.Sprintf("/auth/bookmark/%d", adID))),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Target("this"),
			hx.Swap("outerHTML"),
			g.Attr("onclick", "event.stopPropagation()"),
		},
	})
}

func gridImageNode(adID, count, current int) g.Node {
	if count == 0 {
		return noImage("aspect-[4/3]")
	}
	return ImageNode(adID, count, current, "480w", "aspect-[4/3]", false)
}

func formatListedAge(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "Listed just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "Listed 1 minute ago"
		}
		return fmt.Sprintf("Listed %d minutes ago", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "Listed 1 hour ago"
		}
		return fmt.Sprintf("Listed %d hours ago", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "Listed 1 day ago"
	}
	if days < 7 {
		return fmt.Sprintf("Listed %d days ago", days)
	}
	weeks := days / 7
	if weeks == 1 {
		return "Listed a week ago"
	}
	if days < 30 {
		return fmt.Sprintf("Listed %d weeks ago", weeks)
	}
	months := days / 30
	if months == 1 {
		return "Listed a month ago"
	}
	if days < 365 {
		return fmt.Sprintf("Listed %d months ago", months)
	}
	years := days / 365
	if years == 1 {
		return "Listed a year ago"
	}
	return fmt.Sprintf("Listed %d years ago", years)
}

func newBadge() g.Node {
	return Span(
		Class("px-2 py-0.5 rounded-full bg-zinc-100 text-black text-xs font-medium shadow-sm"),
		g.Text("Just listed"),
	)
}

func newBadgeImageOverlay() g.Node {
	return Div(
		Class("absolute top-[10px] left-[10px] z-10 pointer-events-none"),
		newBadge(),
	)
}

func listedAgeDetailNode(createdAt time.Time) g.Node {
	return Div(
		Class("shrink-0 text-right"),
		g.Text(formatListedAge(createdAt)),
	)
}

func paginationDiv(nextPage int) g.Node {
	return Div(
		hx.Get(fmt.Sprintf("/api/search/?page=%d", nextPage)),
		hx.Trigger("revealed"),
		hx.Target("this"),
		hx.Swap("outerHTML"),
		hx.Include("form"),
	)
}

func adCardMeta(location string, createdAt time.Time) g.Node {
	return Div(
		Class("flex items-center gap-2 text-xs text-zinc-500"),
		g.If(time.Since(createdAt) < 4*time.Hour, newBadge()),
		g.Text(location),
	)
}

func adCardTitle(title, facetLabel string) g.Node {
	return Span(
		Class("min-w-0"),
		g.Text(title),
		g.If(facetLabel != "", Span(Class("text-xs text-zinc-500"), g.Text(" · "+facetLabel))),
	)
}

func priceSpan(priceDisplay string, hasPrice bool) g.Node {
	if !hasPrice {
		return g.Text("")
	}
	return Span(Class("text-green-600 font-semibold"), g.Text(priceDisplay))
}

func AdGridNode(userID, adID, imageCount, nextPage int, priceDisplay, title,
	location, facetLabel, csrfToken string, hasPrice bool, createdAt time.Time,
	active, bookmarked, isLast bool, rockCount int) g.Node {
	class := "flex flex-col cursor-pointer gap-1 py-3"
	if !active {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}

	node := A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		gridImageNode(adID, imageCount, 1),
		Div(
			Class("flex items-center justify-between pt-1"),
			priceSpan(priceDisplay, hasPrice),
			adCardMeta(location, createdAt),
		),
		Div(
			Class("flex items-center gap-2 min-w-0"),
			g.If(local.IsLoggedIn(userID) && bookmarked, BookmarkButton(adID, bookmarked, csrfToken)),
			g.If(rockCount > 0, RockIcons(adID, rockCount)),
			adCardTitle(title, facetLabel),
		),
	)

	if isLast {
		return g.Group([]g.Node{
			node,
			paginationDiv(nextPage),
		})
	}

	return node
}

func AdListNode(userID, adID int, priceDisplay, title, location,
	facetLabel string, hasPrice bool, createdAt time.Time, active, bookmarked bool,
	csrfToken string, isLast bool, nextPage int, rockCount int) g.Node {
	class := "flex flex-wrap items-center justify-between py-2 cursor-pointer"
	if active {
		class += " hover:bg-zinc-50 dark:hover:bg-zinc-800"
	} else {
		class += " bg-red-100 dark:bg-red-900 border border-red-300 dark:border-red-700 rounded-lg"
	}

	node := A(
		Href("/ad/"+strconv.Itoa(adID)),
		Class(class),
		Div(
			Class("flex items-center gap-2 min-w-0"),
			g.If(local.IsLoggedIn(userID) && bookmarked, BookmarkButton(adID, bookmarked, csrfToken)),
			g.If(rockCount > 0, RockIcons(adID, rockCount)),
			adCardTitle(title, facetLabel),
		),
		Div(
			Class("flex items-center gap-2 ml-auto"),
			adCardMeta(location, createdAt),
			priceSpan(priceDisplay, hasPrice),
		),
	)

	if isLast {
		return g.Group([]g.Node{
			node,
			paginationDiv(nextPage),
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

func testWatermark() g.Node {
	return Div(
		Class("absolute top-0 left-0 right-0 bottom-0 flex items-center justify-center pointer-events-none z-50"),
		Div(
			Class("watermark-test text-center whitespace-pre-line"),
			g.Text("TEST\nAD"),
		),
	)
}

func shareButton(adID int) g.Node {
	return iconButton(buttonProps{
		ImageSrc: "/images/share.svg",
		Alt:      "Share Ad",
		Class:    "dark:invert dark:opacity-80",
		Attrs: []g.Node{
			hx.Get(fmt.Sprintf("/api/ad/%d/share", adID)),
			hx.Target("body"),
			hx.Swap("beforeend"),
		},
	})
}

func deleteButton(adID int, csrfToken string) g.Node {
	return iconButton(buttonProps{
		ImageSrc: "/images/trashcan.svg",
		Alt:      "Delete Ad",
		Class:    "dark:invert dark:opacity-80",
		Attrs: []g.Node{
			hx.Delete(fmt.Sprintf("/auth/ad/%d/delete", adID)),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Target("body"),
			hx.Swap("outerHTML"),
			g.Attr("hx-confirm", "Are you sure you want to delete this ad?"),
		},
	})
}

func restoreButton(adID int, csrfToken string) g.Node {
	return iconButton(buttonProps{
		ImageSrc: "/images/restore.svg",
		Alt:      "Restore Ad",
		Class:    "dark:invert dark:opacity-80",
		Attrs: []g.Node{
			hx.Post(fmt.Sprintf("/auth/ad/%d/restore", adID)),
			hx.Headers(fmt.Sprintf(`{"X-Csrf-Token": %q}`, csrfToken)),
			hx.Target("body"),
			hx.Swap("outerHTML"),
			g.Attr("hx-confirm", "Are you sure you want to restore this ad?"),
		},
	})
}

func messageButton(adID int) g.Node {
	return iconButton(buttonProps{
		ImageSrc: "/images/message.svg",
		Alt:      "Message",
		Class:    "dark:invert dark:opacity-80",
		Attrs: []g.Node{
			hx.Get(fmt.Sprintf("/auth/ad/%d/new-conversation", adID)),
			hx.Target("body"),
			hx.Swap("beforeend"),
		},
	})
}

func adButtons(adID, userID, ownerID int, bookmarked, active, reachable bool,
	csrfToken string) g.Node {
	isOwner := local.IsLoggedIn(userID) && userID == ownerID
	return Div(
		Class("flex items-center gap-2"),
		g.If(local.IsLoggedIn(userID), BookmarkButton(adID, bookmarked, csrfToken)),
		g.If(active && local.IsLoggedIn(userID) && reachable && !isOwner, messageButton(adID)),
		g.If(active && isOwner, editButton(adID)),
		g.If(active && isOwner, deleteButton(adID, csrfToken)),
		g.If(!active && isOwner, restoreButton(adID, csrfToken)),
		shareButton(adID),
	)
}

func editButton(adID int) g.Node {
	return iconButton(buttonProps{
		ImageSrc: "/images/edit.svg",
		Alt:      "Edit",
		Class:    "dark:invert dark:opacity-80",
		Attrs: []g.Node{
			g.Attr("onclick", fmt.Sprintf(
				"event.stopPropagation(); window.location.href=%q;",
				fmt.Sprintf("/auth/ad/%d/edit", adID),
			)),
		},
	})
}

func copyButton(path string, copied bool) g.Node {
	var text string
	var attrs []g.Node

	if copied {
		text = "Copied!"
		attrs = []g.Node{
			hx.Get(fmt.Sprintf("/api/ad/share/copy?copied=false&path=%s", url.QueryEscape(path))),
			hx.Target("#copy-link-button"),
			hx.Swap("outerHTML"),
			hx.Trigger("load delay:2s"),
		}
	} else {
		text = "Copy"
		attrs = []g.Node{
			g.Attr("onclick", fmt.Sprintf(`navigator.clipboard.writeText(%q);`, path)),
			hx.Get(fmt.Sprintf("/api/ad/share/copy?copied=true&path=%s", url.QueryEscape(path))),
			hx.Target("#copy-link-button"),
			hx.Swap("outerHTML"),
			hx.Trigger("click"),
		}
	}

	return standardButton(buttonProps{
		Type:  "button",
		ID:    "copy-link-button",
		Class: "flex items-center justify-center gap-2 transition-all duration-200 min-w-[140px]",
		Attrs: attrs,
		Children: []g.Node{
			Div(
				Class("flex items-center justify-center gap-2 whitespace-nowrap"),
				Img(
					Src("/images/copy.svg"),
					Alt("Copy"),
					Class("w-4 h-4"),
				),
				g.Text(text),
			),
		},
	})
}

func CopyButton(path string) g.Node {
	return copyButton(path, false)
}

func CopyButtonCopied(path string) g.Node {
	return copyButton(path, true)
}

func AdShareModal(path string) g.Node {
	return g.Group([]g.Node{
		modalBackdrop("ad-share"),
		Div(
			ID("ad-share-modal"),
			Class("fixed inset-0 flex items-center justify-center z-50 p-8 pointer-events-none"),
			Div(
				Class("bg-white dark:bg-zinc-800 rounded-lg w-full max-w-md shadow-2xl border-2 border-zinc-300 dark:border-zinc-600 flex flex-col pointer-events-auto"),
				Div(
					Class("flex items-center justify-between p-6 border-b border-zinc-200 dark:border-zinc-700 flex-shrink-0"),
					H3(Class("text-xl font-bold text-zinc-900 dark:text-zinc-200"), g.Text("Share Ad")),
					modalClose("ad-share"),
				),
				Div(
					Class("p-6 flex flex-col gap-4"),
					Div(
						Class("flex flex-col gap-2"),
						Label(
							Class("text-sm font-medium text-zinc-700 dark:text-zinc-300"),
							For("ad-link-input"),
							g.Text("Ad Link"),
						),
						Div(
							Class("flex items-center gap-2"),
							Input(
								ID("ad-link-input"),
								Type("text"),
								Value(path),
								g.Attr("readonly", ""),
								g.Attr("onfocus", "this.select();"),
								Class("flex-1 px-3 py-2 border border-zinc-300 dark:border-zinc-600 rounded-md bg-zinc-50 dark:bg-zinc-700 text-sm text-zinc-900 dark:text-zinc-200"),
							),
							copyButton(path, false),
						),
					),
				),
			),
		),
	})
}

func Ad(d AdDetail, userID int, csrfToken string) []g.Node {
	return []g.Node{
		Div(
			Class("flex flex-col relative rounded-lg shadow-lg dark:shadow-xl dark:shadow-zinc-900/50 my-4 mx-2 col-span-full overflow-hidden bg-white dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700"),
			Div(
				Class("relative"),
				g.If(d.ImageCount > 0, ImageNodeWithThumbnails(d.ID, d.ImageCount, 1, "1200w", "aspect-[4/3] w-full", true)),
				g.If(d.ImageCount == 0, noImage("h-32 w-full")),
				g.If(time.Since(d.CreatedAt) < 4*time.Hour, newBadgeImageOverlay()),
			),
			g.If(!d.Active, deletedWatermark()),
			g.If(d.IsTest && d.Active, testWatermark()),
			Div(
				Class("p-6 flex flex-col bg-white dark:bg-zinc-800"),
				Div(
					Class("flex items-center justify-between min-w-0"),
					Div(
						Class("flex items-center gap-2 min-w-0"),
						g.If(d.RockCount > 0, RockIcons(d.ID, d.RockCount)),
						adCardTitle(d.Title, d.FacetLabel),
					),
					adButtons(d.ID, userID, d.OwnerID, d.Bookmarked, d.Active, d.Reachable, csrfToken),
				),
				Div(
					Class("flex items-center gap-2 min-w-0"),
					priceSpan(d.PriceDisplay, d.HasPrice),
					Div(
						Class("flex flex-1 items-center justify-between gap-2 text-xs text-zinc-500 min-w-0"),
						Span(Class("min-w-0"), g.Text(d.Location)),
						listedAgeDetailNode(d.CreatedAt),
					),
				),
				descriptionDisplay(
					d.ID,
					d.DescriptionOriginal,
					d.FacetDetails,
					d.Tags,
					d.DescriptionHistory,
				),
			),
		),
	}
}

func descriptionDisplay(adID int, original string, facetDetails []string,
	tags []string, history []AdHistoryEntry) g.Node {
	var nodes []g.Node
	if len(facetDetails) > 0 {
		nodes = append(nodes, adFacetList(facetDetails))
	}
	nodes = append(nodes, Div(Class("whitespace-pre-wrap"), g.Text(original)))
	if len(tags) > 0 {
		nodes = append(nodes, adPills(tags, "mt-3"))
	}
	if len(history) > 0 {
		entryNodes := make([]g.Node, len(history))
		for i, e := range history {
			entryNodes[i] = descriptionHistoryEntry(adID, e)
		}
		nodes = append(nodes, Div(
			Class("mt-4 pt-4 border-t border-zinc-200 dark:border-zinc-600 space-y-4"),
			g.Group(entryNodes),
		))
	}
	return Div(Class("text-base mt-4"), g.Group(nodes))
}

func descriptionHistoryEntry(adID int, e AdHistoryEntry) g.Node {
	imageNodes := make([]g.Node, len(e.ImageIndices))
	for i, idx := range e.ImageIndices {
		imageNodes[i] = Img(
			Src(fmt.Sprintf("/ad/%d/image/%d/%s", adID, idx, "160w")),
			Alt(fmt.Sprintf("Added image %d", idx)),
			Class("w-16 h-16 object-cover rounded border "+
				"border-zinc-200 dark:border-zinc-600"),
		)
	}
	return Div(
		Class("text-sm text-blue-700 dark:text-blue-300"),
		Div(
			Class("flex items-start gap-2"),
			Img(
				Src("/images/edit.svg"),
				Alt("Edit"),
				Class("w-4 h-4 mt-0.5 shrink-0 dark:invert dark:opacity-80"),
			),
			Div(
				Class("min-w-0"),
				Div(Class("font-medium"), g.Text(e.Header)),
				g.If(e.Body != "",
					Div(
						Class("whitespace-pre-wrap mt-1 "+
							"text-blue-600/90 dark:text-blue-200/90"),
						g.Text(e.Body),
					),
				),
				g.If(len(e.ImageIndices) > 0,
					Div(
						Class("flex flex-wrap gap-2 mt-2"),
						g.Group(imageNodes),
					),
				),
			),
		),
	)
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
				Class("text-3xl font-bold mb-4"),
				g.Text("Ad Deleted"),
			),
			P(
				Class("text-lg text-zinc-600 dark:text-zinc-400 mb-8"),
				g.Text("This ad has been deleted by the owner and is no longer available."),
			),
			standardButton(buttonProps{
				Text: "Back to Home",
				Href: "/",
			}),
		),
	}
}

func imagesField(maxImagesPerAd int) g.Node {
	return imageUploadField(maxImagesPerAd, 0, false)
}

func editImagesField(adID, existingCount, maxImagesPerAd int) g.Node {
	if existingCount == 0 {
		return imageUploadField(maxImagesPerAd, 0, false)
	}

	existingNodes := make([]g.Node, existingCount)
	for i := 1; i <= existingCount; i++ {
		existingNodes[i-1] = Img(
			Src(fmt.Sprintf("/ad/%d/image/%d/%s", adID, i, "160w")),
			Alt(fmt.Sprintf("Image %d", i)),
			Class("object-cover rounded w-[90px] h-[90px]"),
		)
	}

	nodes := []g.Node{
		label("Images"),
		Div(
			Class("text-sm text-zinc-500 dark:text-zinc-400"),
			g.Text("Original images cannot be changed."),
		),
		Div(
			ID("existing-image-preview"),
			Class("flex flex-row flex-wrap gap-2"),
			g.Group(existingNodes),
		),
	}

	if existingCount >= maxImagesPerAd {
		nodes = append(nodes, Div(
			Class("text-sm text-zinc-500 dark:text-zinc-400"),
			g.Text(fmt.Sprintf(
				"Maximum of %d images reached.", maxImagesPerAd,
			)),
		))
	} else {
		nodes = append(nodes,
			Div(
				Class("text-sm text-zinc-500 dark:text-zinc-400"),
				g.Text("Add more images (optional)"),
			),
			imageUploadControls(maxImagesPerAd, existingCount, true),
		)
	}

	return Div(Class("space-y-2"), g.Group(nodes))
}

func imageUploadField(maxImagesPerAd, existingCount int, appendMode bool) g.Node {
	return Div(
		Class("space-y-2"),
		label("Images"),
		imageUploadControls(maxImagesPerAd, existingCount, appendMode),
	)
}

func imageUploadControls(maxImagesPerAd, existingCount int,
	appendMode bool) g.Node {
	appendFlag := "false"
	if appendMode {
		appendFlag = "true"
	}
	return Div(
		Input(
			Type("file"),
			ID("images"),
			Name("images"),
			Class("hidden"),
			g.Attr("accept", "image/*"),
			g.Attr("multiple"),
			g.Attr("onchange", "previewImages(this)"),
		),
		Div(
			ID("upload-area"),
			Class("border border-zinc-300 dark:border-zinc-600 rounded p-6 "+
				"hover:border-blue-400 hover:bg-blue-50 dark:hover:bg-zinc-800 "+
				"transition-colors duration-200 cursor-pointer"),
			g.Attr("onclick", "handleUploadClick()"),
			g.Attr("ondragover", "event.preventDefault(); "+
				"this.classList.add('border-blue-400', 'bg-blue-50')"),
			g.Attr("ondragleave", "this.classList.remove('border-blue-400', "+
				"'bg-blue-50')"),
			g.Attr("ondrop", "event.preventDefault(); "+
				"this.classList.remove('border-blue-400', 'bg-blue-50'); "+
				"handleDrop(event)"),
			Div(
				ID("upload-content"),
				Class("flex flex-col items-center space-y-4"),
				Div(
					Class("flex flex-col items-center space-y-2"),
					Div(
						Class("w-12 h-12 bg-blue-100 dark:bg-zinc-700 "+
							"rounded-full flex items-center justify-center"),
						g.Raw(`<svg class="w-6 h-6 text-blue-600 dark:text-blue-400" `+
							`fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" `+
							`stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
						</svg>`),
					),
					Div(
						Class("text-lg font-medium text-zinc-700 dark:text-zinc-300"),
						g.Text("Upload Images"),
					),
					Div(
						Class("text-sm text-zinc-500 dark:text-zinc-400"),
						g.Text("Click to browse or drag and drop"),
					),
				),
			),
			Div(
				ID("image-preview"),
				Class("hidden image-preview flex flex-row flex-wrap "+
					"gap-2 justify-center mt-4"),
			),
		),
		g.Raw(fmt.Sprintf(
			`<script>const MAX_IMAGES_PER_AD = %d;`+
				`const EXISTING_IMAGE_COUNT = %d;`+
				`const IMAGE_APPEND_MODE = %s;</script>`,
			maxImagesPerAd, existingCount, appendFlag,
		)),
		g.Raw(`<script src="/js/image-preview.js" defer></script>`),
	)
}

func adFacetList(items []string) g.Node {
	listItems := make([]g.Node, len(items))
	for i, s := range items {
		listItems[i] = Li(g.Text(s))
	}
	return Ul(
		Class("list-disc list-inside mb-3 text-sm text-zinc-700 dark:text-zinc-300"),
		g.Group(listItems),
	)
}

func adPills(labels []string, spacing string) g.Node {
	nodes := make([]g.Node, len(labels))
	for i, s := range labels {
		nodes[i] = Span(
			Class("inline-block px-3 py-1 rounded-full border border-zinc-300 dark:border-zinc-600 text-sm text-zinc-700 dark:text-zinc-300"),
			g.Text(s),
		)
	}
	return Div(
		Class("flex flex-wrap gap-2 "+spacing),
		g.Group(nodes),
	)
}

func newAdForm(fields g.Node) g.Node {
	cfg := uiads.NewFormConfig(uiads.AdFormConfig{}.Defaults)
	return adForm(cfg, fields)
}

func adForm(cfg uiads.AdFormConfig, fields g.Node) g.Node {
	formAttrs := []g.Node{
		Class("space-y-8 mt-8"),
		ID(cfg.FormID),
		g.Attr("novalidate", ""),
		hx.Post(cfg.PostURL),
		hx.Swap("none"),
	}
	if cfg.Mode == uiads.AdFormCreate ||
		(cfg.Mode == uiads.AdFormEdit &&
			cfg.Values.ImageCount < config.MaxImagesPerAd) {
		formAttrs = append(formAttrs, hx.Encoding("multipart/form-data"))
	}

	children := []g.Node{fields}
	if cfg.Mode == uiads.AdFormCreate {
		children = append(children, imagesField(config.MaxImagesPerAd))
	} else if cfg.Mode == uiads.AdFormEdit {
		children = append(children, editImagesField(
			cfg.AdID, cfg.Values.ImageCount, config.MaxImagesPerAd,
		))
	}
	children = append(children,
		Div(
			Class("flex items-center gap-4"),
			standardButton(buttonProps{
				Type: "submit",
				Text: cfg.SubmitLabel,
			}),
			ErrorDiv(""),
		),
	)

	return Form(append(formAttrs, g.Group(children))...)
}

func NewAd(category CategoryOption, fields g.Node) []g.Node {
	return []g.Node{
		Div(Class("mb-4"), categoryButton(category, "/auth/ad/new")),
		pageTitle("Create New Ad"),
		newAdForm(fields),
	}
}

func EditAd(category CategoryOption, cfg uiads.AdFormConfig,
	fields g.Node) []g.Node {
	imagePath := "/images/category/" + category.ImageFile
	return []g.Node{
		Div(
			Class("py-2 px-5 flex items-center gap-2 rounded-full border-2 mb-4 "+
				"border-zinc-300 bg-zinc-100 dark:bg-zinc-800 dark:border-zinc-600"),
			Img(
				Src(imagePath),
				Alt("Category icon"),
				Class("w-6 h-6 dark:invert dark:opacity-80"),
			),
			Span(g.Text(category.Name)),
		),
		pageTitle("Edit Ad"),
		adForm(cfg, fields),
	}
}

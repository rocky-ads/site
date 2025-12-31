package ui

import (
	"fmt"
	"math"
	"strconv"

	g "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/html"
)

func GenerateSVG(adID, imageID int, size string) g.Node {
	var width int

	switch size {
	case "160w":
		width = 160
	case "480w":
		width = 480
	case "1200w":
		width = 1200
	default:
		width = 480 // fallback
	}

	// 4:3 aspect ratio → height = width * 3 / 4
	height := width * 3 / 4

	// Scale font size based on the smaller dimension (minimum 12px), then halve it
	fontSize := int(math.Max(12, math.Min(float64(width), float64(height))/6)) / 2
	fontSizeStr := strconv.Itoa(fontSize)
	lineHeight := int(float64(fontSize) * 1.2)

	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("width", strconv.Itoa(width)),
		g.Attr("height", strconv.Itoa(height)),
		g.Attr("viewBox", fmt.Sprintf("0 0 %d %d", width, height)),

		// Define checkerboard pattern rotated 45 degrees
		g.El("defs",
			g.El("pattern",
				g.Attr("id", fmt.Sprintf("checkerboard-%d-%d", adID, imageID)),
				g.Attr("patternUnits", "userSpaceOnUse"),
				g.Attr("width", "40"),
				g.Attr("height", "40"),
				g.Attr("patternTransform", "rotate(45)"),
				g.El("rect",
					g.Attr("width", "20"),
					g.Attr("height", "20"),
					g.Attr("fill", "#9ca3af"), // gray-400 - lighter squares
				),
				g.El("rect",
					g.Attr("x", "20"),
					g.Attr("y", "20"),
					g.Attr("width", "20"),
					g.Attr("height", "20"),
					g.Attr("fill", "#9ca3af"), // gray-400 - lighter squares
				),
				g.El("rect",
					g.Attr("x", "20"),
					g.Attr("y", "0"),
					g.Attr("width", "20"),
					g.Attr("height", "20"),
					g.Attr("fill", "#6b7280"), // gray-500 - darker squares
				),
				g.El("rect",
					g.Attr("x", "0"),
					g.Attr("y", "20"),
					g.Attr("width", "20"),
					g.Attr("height", "20"),
					g.Attr("fill", "#6b7280"), // gray-500 - darker squares
				),
			),
		),

		// Background with checkerboard pattern
		g.El("rect",
			g.Attr("width", "100%"),
			g.Attr("height", "100%"),
			g.Attr("fill", fmt.Sprintf("url(#checkerboard-%d-%d)", adID, imageID)),
		),

		// Centered image text
		g.El("text",
			g.Attr("x", "50%"),
			g.Attr("y", fmt.Sprintf("%d", height/2-lineHeight)),
			g.Attr("font-size", fontSizeStr),
			g.Attr("font-family", "sans-serif"),
			g.Attr("fill", "#ffffff"), // white text for good contrast
			g.Attr("text-anchor", "middle"),
			g.Attr("dominant-baseline", "middle"),
			g.El("tspan",
				g.Attr("x", "50%"),
				g.Attr("dy", "0"),
				g.Text(fmt.Sprintf("ad: %d, image: %d", adID, imageID)),
			),
			g.El("tspan",
				g.Attr("x", "50%"),
				g.Attr("dy", fmt.Sprintf("%d", lineHeight)),
				g.Text(fmt.Sprintf("size: %s", size)),
			),
			g.El("tspan",
				g.Attr("x", "50%"),
				g.Attr("dy", fmt.Sprintf("%d", lineHeight)),
				g.Text("---"),
			),
			g.El("tspan",
				g.Attr("x", "50%"),
				g.Attr("dy", fmt.Sprintf("%d", lineHeight)),
				g.Text("<missing>"),
			),
		),
	)
}

func noImage(heightClass string) g.Node {
	return Div(
		Class(fmt.Sprintf("rounded-md w-full %s flex items-center justify-center border-2 border-dotted border-zinc-300 dark:border-zinc-600", heightClass)),
		g.Text("No Image"),
	)
}

func imageCount(current, count int) g.Node {
	return Span(
		Class("absolute bottom-2 right-2 bg-black/50 text-white text-xs px-2 py-1 rounded-full"),
		g.Text(fmt.Sprintf("%d/%d", current, count)),
	)
}

func imageNav(adID, current, count int, size, heightClass string, clickable bool) g.Node {
	if count <= 1 {
		return g.Group([]g.Node{}) // Return empty group if count is 0 or 1
	}
	prevIdx := (current-2+count)%count + 1
	nextIdx := current%count + 1
	containerID := fmt.Sprintf("image-%d", adID)

	return g.Group([]g.Node{
		// Left button
		Button(
			Type("button"),
			Class("absolute left-2 top-1/2 transform -translate-y-1/2 bg-white/50 rounded-full w-10 h-10 flex items-center justify-center shadow-lg hover:bg-white/60 focus:outline-none cursor-pointer z-20 opacity-100 md:opacity-0 md:group-hover:opacity-100 md:transition-opacity"),
			hx.Get(fmt.Sprintf("/api/image-nav/%d?index=%d&count=%d&size=%s&heightClass=%s&clickable=%v",
				adID, prevIdx, count, size, heightClass, clickable)),
			hx.Target(fmt.Sprintf("#%s", containerID)),
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
			hx.Get(fmt.Sprintf("/api/image-nav/%d?index=%d&count=%d&size=%s&heightClass=%s&clickable=%v",
				adID, nextIdx, count, size, heightClass, clickable)),
			hx.Target(fmt.Sprintf("#%s", containerID)),
			hx.Swap("outerHTML"),
			g.Attr("onclick", "event.stopPropagation()"),
			Img(
				Class("w-6 h-6"),
				Src("/images/right.svg"),
			),
		),
	})
}

func ImageNode(adID, count, current int, size, heightClass string, clickable bool) g.Node {
	containerID := fmt.Sprintf("image-%d", adID)
	imgElement := Img(
		Class(fmt.Sprintf("rounded-t-md w-full %s object-cover", heightClass)),
		Src(fmt.Sprintf("/ad/%d/image/%d/%s", adID, current, size)),
		g.Attr("loading", "lazy"),
	)

	var imageWrapper g.Node
	if clickable {
		imageWrapper = Div(
			hx.Get(fmt.Sprintf("/api/image-full/%d?index=%d&count=%d&size=%s", adID, current, count, size)),
			hx.Target("body"),
			hx.Swap("beforeend"),
			Class("cursor-pointer"),
			imgElement,
		)
	} else {
		imageWrapper = imgElement
	}

	return Div(
		ID(containerID),
		Class("relative group"),
		imageWrapper,
		g.If(count > 1, imageCount(current, count)),
		g.If(count > 1, imageNav(adID, current, count, size, heightClass, clickable)),
	)
}

func imageFullScreenContent(adID, current, count int, size string) g.Node {
	return Div(
		Class("relative w-full h-full flex items-center justify-center pointer-events-auto"),
		// Close button
		Button(
			Type("button"),
			Class("absolute top-4 right-4 bg-white/90 hover:bg-white rounded-full w-10 h-10 flex items-center justify-center shadow-lg focus:outline-none cursor-pointer z-50"),
			hx.Get("/api/modal-remove/image-fullscreen"),
			hx.Swap("none"),
			Img(
				Src("/images/close.svg"),
				Alt("Close"),
				Class("w-6 h-6"),
			),
		),
		// Image container
		Div(
			Class("relative w-full h-full flex items-center justify-center"),
			Img(
				Class("max-w-full max-h-full object-contain"),
				Src(fmt.Sprintf("/ad/%d/image/%d/%s", adID, current, size)),
				Alt(fmt.Sprintf("Image %d of %d", current, count)),
			),
			// Navigation buttons (only if multiple images)
			g.If(count > 1, imageFullScreenNav(adID, current, count, size)),
			// Image counter
			g.If(count > 1, imageFullScreenCount(current, count)),
		),
	)
}

func ImageFullScreen(adID, current, count int, size string) g.Node {
	return g.Group([]g.Node{
		Div(
			ID("image-fullscreen-modal-backdrop"),
			Class("fixed inset-0 bg-black/90 z-40"),
			hx.Get("/api/modal-remove/image-fullscreen"),
			hx.Swap("none"),
			hx.Trigger("click"),
		),
		Div(
			ID("image-fullscreen-modal"),
			Class("fixed inset-0 flex items-center justify-center z-50 p-4 pointer-events-none"),
			imageFullScreenContent(adID, current, count, size),
		),
	})
}

func ImageFullScreenUpdate(adID, current, count int, size string) g.Node {
	return g.Group([]g.Node{
		Div(
			ID("image-fullscreen-modal-backdrop"),
			hx.SwapOOB("true"),
			Class("fixed inset-0 bg-black/90 z-40"),
			hx.Get("/api/modal-remove/image-fullscreen"),
			hx.Swap("none"),
			hx.Trigger("click"),
		),
		Div(
			ID("image-fullscreen-modal"),
			hx.SwapOOB("true"),
			Class("fixed inset-0 flex items-center justify-center z-50 p-4 pointer-events-none"),
			imageFullScreenContent(adID, current, count, size),
		),
	})
}

func imageFullScreenNav(adID, current, count int, size string) g.Node {
	prevIdx := (current-2+count)%count + 1
	nextIdx := current%count + 1

	return g.Group([]g.Node{
		// Left button
		Button(
			Type("button"),
			Class("absolute left-4 top-1/2 transform -translate-y-1/2 bg-white/90 hover:bg-white rounded-full w-12 h-12 flex items-center justify-center shadow-lg focus:outline-none cursor-pointer z-40"),
			hx.Get(fmt.Sprintf("/api/image-full/%d?index=%d&count=%d&size=%s&update=true", adID, prevIdx, count, size)),
			hx.Target("body"),
			hx.Swap("none"),
			Img(
				Class("w-6 h-6"),
				Src("/images/left.svg"),
			),
		),
		// Right button
		Button(
			Type("button"),
			Class("absolute right-4 top-1/2 transform -translate-y-1/2 bg-white/90 hover:bg-white rounded-full w-12 h-12 flex items-center justify-center shadow-lg focus:outline-none cursor-pointer z-40"),
			hx.Get(fmt.Sprintf("/api/image-full/%d?index=%d&count=%d&size=%s&update=true", adID, nextIdx, count, size)),
			hx.Target("body"),
			hx.Swap("none"),
			Img(
				Class("w-6 h-6"),
				Src("/images/right.svg"),
			),
		),
	})
}

func imageFullScreenCount(current, count int) g.Node {
	return Div(
		Class("absolute bottom-4 left-1/2 transform -translate-x-1/2 bg-black/70 text-white text-sm px-4 py-2 rounded-full"),
		g.Text(fmt.Sprintf("%d / %d", current, count)),
	)
}

func ImageThumbnails(adID, current, count int, size, heightClass string, clickable bool) g.Node {
	if count <= 1 {
		return g.Group([]g.Node{}) // Return empty group if count is 0 or 1
	}
	containerID := fmt.Sprintf("image-%d", adID)

	thumbnails := make([]g.Node, count)
	for i := 1; i <= count; i++ {
		isCurrent := i == current
		borderClass := "border-2 "
		if isCurrent {
			borderClass += "border-blue-500"
		} else {
			borderClass += "border-transparent hover:border-zinc-300 dark:hover:border-zinc-600"
		}

		thumbnailImg := Img(
			Class("w-full h-full object-cover rounded"),
			Src(fmt.Sprintf("/ad/%d/image/%d/%s", adID, i, size)),
			g.Attr("loading", "lazy"),
		)

		var thumbnailWrapper g.Node
		if clickable {
			thumbnailWrapper = Button(
				Type("button"),
				Class(fmt.Sprintf("flex-shrink-0 w-24 aspect-[4/3] rounded overflow-hidden cursor-pointer %s transition-colors", borderClass)),
				hx.Get(fmt.Sprintf("/api/image-nav/%d?index=%d&count=%d&size=%s&heightClass=%s&clickable=%v",
					adID, i, count, "1200w", heightClass, clickable)),
				hx.Target(fmt.Sprintf("#%s", containerID)),
				hx.Swap("outerHTML"),
				g.Attr("onclick", "event.stopPropagation()"),
				thumbnailImg,
			)
		} else {
			thumbnailWrapper = Div(
				Class(fmt.Sprintf("flex-shrink-0 w-24 aspect-[4/3] rounded overflow-hidden %s", borderClass)),
				thumbnailImg,
			)
		}

		thumbnails[i-1] = thumbnailWrapper
	}

	return Div(
		Class("flex flex-wrap gap-2 px-4 py-2 justify-center"),
		g.Group(thumbnails),
	)
}

func ImageNodeWithThumbnails(adID, count, current int, size, heightClass string, clickable bool) g.Node {
	containerID := fmt.Sprintf("image-%d", adID)
	imgElement := Img(
		Class(fmt.Sprintf("rounded-t-md w-full %s object-cover", heightClass)),
		Src(fmt.Sprintf("/ad/%d/image/%d/%s", adID, current, size)),
		g.Attr("loading", "lazy"),
	)

	var imageWrapper g.Node
	if clickable {
		imageWrapper = Div(
			hx.Get(fmt.Sprintf("/api/image-full/%d?index=%d&count=%d&size=%s", adID, current, count, size)),
			hx.Target("body"),
			hx.Swap("beforeend"),
			Class("cursor-pointer"),
			imgElement,
		)
	} else {
		imageWrapper = imgElement
	}

	return Div(
		ID(containerID),
		Class("relative group"),
		imageWrapper,
		g.If(count > 1, imageNav(adID, current, count, size, heightClass, clickable)),
		g.If(count > 1, ImageThumbnails(adID, current, count, "160w", heightClass, clickable)),
	)
}

package ui

import (
	"fmt"
	"math"
	"strconv"

	g "maragu.dev/gomponents"
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

	// Scale font size based on the smaller dimension (minimum 12px)
	fontSize := int(math.Max(12, math.Min(float64(width), float64(height))/6))
	fontSizeStr := strconv.Itoa(fontSize)
	lineHeight := int(float64(fontSize) * 1.2)

	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("width", strconv.Itoa(width)),
		g.Attr("height", strconv.Itoa(height)),
		g.Attr("viewBox", fmt.Sprintf("0 0 %d %d", width, height)),

		// Solid background
		g.El("rect",
			g.Attr("width", "100%"),
			g.Attr("height", "100%"),
			g.Attr("fill", "#3498db"), // blue
		),

		// Centered image text
		g.El("text",
			g.Attr("x", "50%"),
			g.Attr("y", fmt.Sprintf("%d", height/2-lineHeight)),
			g.Attr("font-size", fontSizeStr),
			g.Attr("font-family", "sans-serif"),
			g.Attr("fill", "#ffffff"),
			g.Attr("text-anchor", "middle"),
			g.Attr("dominant-baseline", "middle"),
			g.El("tspan",
				g.Attr("x", "50%"),
				g.Attr("dy", "0"),
				g.Text(fmt.Sprintf("%d", adID)),
			),
			g.El("tspan",
				g.Attr("x", "50%"),
				g.Attr("dy", fmt.Sprintf("%d", lineHeight)),
				g.Text(fmt.Sprintf("%d", imageID)),
			),
			g.El("tspan",
				g.Attr("x", "50%"),
				g.Attr("dy", fmt.Sprintf("%d", lineHeight)),
				g.Text(fmt.Sprintf("%dx%d", width, height)),
			),
		),
	)
}

package businesscard

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"path/filepath"

	"github.com/skip2/go-qrcode"
)

const (
	printDPI     = 300
	trimWidthIn  = 3.5
	trimHeightIn = 2.0
	bleedIn      = 0.125
)

var brandBlack = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}

// PrintCardOptions configures flat business card image generation.
type PrintCardOptions struct {
	Category   Category
	Host       string
	ImagesRoot string
}

// RenderPrintCard draws a print-ready business card image at 300 DPI with
// standard bleed margins for professional printing.
func RenderPrintCard(opts PrintCardOptions) (*image.RGBA, error) {
	if opts.Host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if opts.ImagesRoot == "" {
		opts.ImagesRoot = "static/images/category"
	}

	totalW := inchesToPx(trimWidthIn + 2*bleedIn)
	totalH := inchesToPx(trimHeightIn + 2*bleedIn)
	trimW := inchesToPx(trimWidthIn)
	trimH := inchesToPx(trimHeightIn)
	bleedPx := inchesToPx(bleedIn)

	canvas := image.NewRGBA(image.Rect(0, 0, totalW, totalH))

	trim := image.Rect(bleedPx, bleedPx, bleedPx+trimW, bleedPx+trimH)
	cx := trim.Min.X + trimW/2

	const (
		titleSize    = 26.0
		categorySize = 18.0
		taglineSize  = 11.0
		iconSize     = 100
		iconGap      = 20
		qrSize       = 300
		taglineGap   = 12
	)

	const (
		titleText   = "Rocky Ads"
		taglineText = "Scan to browse"
	)

	titleH, err := textHeight(titleText, titleSize)
	if err != nil {
		return nil, err
	}
	taglineH, err := textHeight(taglineText, taglineSize)
	if err != nil {
		return nil, err
	}

	stackH := titleH + iconSize + qrSize + taglineGap + taglineH
	outerMargin := (trimH - stackH) / 2
	if outerMargin < 0 {
		outerMargin = 0
	}

	titleTop := trim.Min.Y + outerMargin
	titleBaseline, err := textBaselineFromTop(titleText, titleTop, titleSize)
	if err != nil {
		return nil, err
	}
	titleBottom, err := textBottomY(titleText, titleBaseline, titleSize)
	if err != nil {
		return nil, err
	}

	taglineBottom := trim.Max.Y - outerMargin
	taglineBaseline, err := textBaselineFromTop(
		taglineText, taglineBottom-taglineH, taglineSize,
	)
	if err != nil {
		return nil, err
	}
	taglineTop := taglineBottom - taglineH

	qrY := taglineTop - taglineGap - qrSize

	if err := drawCenteredText(
		canvas, titleText, cx, titleBaseline, titleSize, brandBlack,
	); err != nil {
		return nil, err
	}

	categoryRowCenterY := (titleBottom + qrY) / 2
	categoryBaseline, err := textBaselineForCenterY(
		opts.Category.Name, categoryRowCenterY, categorySize,
	)
	if err != nil {
		return nil, err
	}

	categoryWidth, err := textWidth(opts.Category.Name, categorySize)
	if err != nil {
		return nil, err
	}
	groupWidth := iconSize + iconGap + categoryWidth
	groupLeft := cx - groupWidth/2

	iconPath := filepath.Join(opts.ImagesRoot, opts.Category.ImageFile)
	icon, err := renderSVGRGBA(iconPath, iconSize, brandBlack)
	if err != nil {
		return nil, fmt.Errorf("render icon %s: %w", iconPath, err)
	}

	iconX := groupLeft
	iconY := categoryRowCenterY - iconSize/2
	drawAt(canvas, icon, iconX, iconY)

	categoryX := groupLeft + iconSize + iconGap
	if err := drawLeftText(
		canvas, opts.Category.Name, categoryX, categoryBaseline,
		categorySize, brandBlack,
	); err != nil {
		return nil, err
	}

	if err := drawCenteredText(
		canvas, taglineText, cx, taglineBaseline, taglineSize, brandBlack,
	); err != nil {
		return nil, err
	}

	qrURL := CategoryURL(opts.Host, opts.Category.ID)
	qr, err := renderQRRGBA(qrURL, qrSize, brandBlack)
	if err != nil {
		return nil, err
	}
	qrX := cx - qrSize/2
	drawAt(canvas, qr, qrX, qrY)

	return canvas, nil
}

// WritePNG encodes img as a PNG with alpha transparency.
func WritePNG(w io.Writer, img image.Image) error {
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(w, img)
}

func inchesToPx(inches float64) int {
	return int(inches*printDPI + 0.5)
}

func drawAt(dst *image.RGBA, src image.Image, x, y int) {
	draw.Draw(dst, image.Rect(x, y, x+src.Bounds().Dx(), y+src.Bounds().Dy()),
		src, src.Bounds().Min, draw.Over)
}

func renderQRRGBA(url string, size int, fg color.Color) (*image.RGBA, error) {
	code, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	img := code.Image(size)
	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if c.Y < 128 {
				rgba.Set(x, y, fg)
			}
		}
	}
	return rgba, nil
}

package businesscard

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"os"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func opentypeFace(size, dpi float64) (font.Face, error) {
	f, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

func drawCenteredText(
	dst *image.RGBA,
	text string,
	centerX, baselineY int,
	size float64,
	clr color.Color,
) error {
	face, err := opentypeFace(size, printDPI)
	if err != nil {
		return err
	}
	defer face.Close()

	bounds, _ := font.BoundString(face, text)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	x := centerX - textWidth/2

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(clr),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(baselineY)},
	}
	d.DrawString(text)
	return nil
}

func drawLeftText(
	dst *image.RGBA,
	text string,
	x, baselineY int,
	size float64,
	clr color.Color,
) error {
	face, err := opentypeFace(size, printDPI)
	if err != nil {
		return err
	}
	defer face.Close()

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(clr),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(baselineY)},
	}
	d.DrawString(text)
	return nil
}

func textWidth(text string, size float64) (int, error) {
	face, err := opentypeFace(size, printDPI)
	if err != nil {
		return 0, err
	}
	defer face.Close()

	bounds, _ := font.BoundString(face, text)
	return (bounds.Max.X - bounds.Min.X).Ceil(), nil
}

func textBaselineForCenterY(text string, centerY int, size float64) (int, error) {
	face, err := opentypeFace(size, printDPI)
	if err != nil {
		return 0, err
	}
	defer face.Close()

	bounds, _ := font.BoundString(face, text)
	midOffset := (bounds.Min.Y + bounds.Max.Y).Ceil() / 2
	return centerY - midOffset, nil
}

func textBottomY(text string, baselineY int, size float64) (int, error) {
	face, err := opentypeFace(size, printDPI)
	if err != nil {
		return 0, err
	}
	defer face.Close()

	bounds, _ := font.BoundString(face, text)
	return baselineY + bounds.Max.Y.Ceil(), nil
}

func textHeight(text string, size float64) (int, error) {
	face, err := opentypeFace(size, printDPI)
	if err != nil {
		return 0, err
	}
	defer face.Close()

	bounds, _ := font.BoundString(face, text)
	return bounds.Max.Y.Ceil() - bounds.Min.Y.Ceil(), nil
}

func textBaselineFromTop(text string, topY int, size float64) (int, error) {
	face, err := opentypeFace(size, printDPI)
	if err != nil {
		return 0, err
	}
	defer face.Close()

	bounds, _ := font.BoundString(face, text)
	return topY - bounds.Min.Y.Ceil(), nil
}

func renderSVGRGBA(path string, sizePx int, tint color.Color) (*image.RGBA, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	icon, err := oksvg.ReadIconStream(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	scale := float64(sizePx) / icon.ViewBox.W
	if icon.ViewBox.W == 0 {
		scale = 1
	}
	icon.Transform = rasterx.Identity.Scale(scale, scale).
		Translate(-icon.ViewBox.X, -icon.ViewBox.Y)

	rgba := image.NewRGBA(image.Rect(0, 0, sizePx, sizePx))
	draw.Draw(rgba, rgba.Bounds(), image.Transparent, image.Point{}, draw.Src)
	scanner := rasterx.NewScannerGV(sizePx, sizePx, rgba, rgba.Bounds())
	dasher := rasterx.NewDasher(sizePx, sizePx, scanner)
	icon.Draw(dasher, 1.0)

	out := image.NewRGBA(image.Rect(0, 0, sizePx, sizePx))
	for y := 0; y < sizePx; y++ {
		for x := 0; x < sizePx; x++ {
			_, _, _, a := rgba.At(x, y).RGBA()
			if a > 0x8000 {
				out.Set(x, y, tint)
			}
		}
	}
	return out, nil
}

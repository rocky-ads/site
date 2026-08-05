package businesscard

import (
	"bytes"
	"image"
	"os"
	"testing"

	"github.com/srwiley/oksvg"
)

func TestSVGRGBAHasInk(t *testing.T) {
	path := ResolveRepoPath("static/images/category/bicycle.svg")
	img, err := renderSVGRGBA(path, 96, brandText)
	if err != nil {
		t.Fatalf("renderSVGRGBA: %v", err)
	}
	if countOpaquePixels(img) < 50 {
		t.Fatal("expected SVG ink pixels")
	}
}

func TestSVGIconParsesPaths(t *testing.T) {
	data, err := os.ReadFile(ResolveRepoPath("static/images/category/bicycle.svg"))
	if err != nil {
		t.Fatalf("read svg: %v", err)
	}
	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ReadIconStream: %v", err)
	}
	if len(icon.SVGPaths) == 0 {
		t.Fatal("expected SVG paths")
	}
}

func countOpaquePixels(img *image.RGBA) int {
	n := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A > 128 {
				n++
			}
		}
	}
	return n
}

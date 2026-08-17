package imgconv

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/chai2010/webp"
)

func onePx() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 200, A: 255})
	return img
}

func TestToJPEGPassthrough(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, onePx(), &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	in := buf.Bytes()
	out, err := ToJPEG(in, DefaultQuality)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(in, out) {
		t.Fatal("jpeg should pass through unchanged")
	}
}

func TestToJPEGFromPNG(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, onePx()); err != nil {
		t.Fatal(err)
	}
	out, err := ToJPEG(buf.Bytes(), DefaultQuality)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 3 || out[0] != 0xff || out[1] != 0xd8 {
		t.Fatalf("expected jpeg, got %d bytes", len(out))
	}
}

func TestToJPEGFromWebP(t *testing.T) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, onePx(), &webp.Options{
		Lossless: false, Quality: 80,
	}); err != nil {
		t.Fatal(err)
	}
	out, err := ToJPEG(buf.Bytes(), DefaultQuality)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 3 || out[0] != 0xff || out[1] != 0xd8 {
		t.Fatalf("expected jpeg, got %d bytes", len(out))
	}
}

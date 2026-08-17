package imgconv

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// 2x2 lossy WebP (legacy objects we still convert).
var tinyWebP = []byte{
	0x52, 0x49, 0x46, 0x46, 0x46, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50,
	0x56, 0x50, 0x38, 0x20, 0x3a, 0x00, 0x00, 0x00, 0x50, 0x02, 0x00, 0x9d,
	0x01, 0x2a, 0x02, 0x00, 0x02, 0x00, 0x01, 0x40, 0x26, 0x25, 0xa0, 0x02,
	0x74, 0xba, 0x01, 0xf8, 0x00, 0x03, 0x21, 0x33, 0xee, 0xe6, 0x00, 0x00,
	0xfe, 0xf9, 0x53, 0x5b, 0xaf, 0xa6, 0xb7, 0x24, 0x7e, 0x5a, 0x71, 0xff,
	0xe4, 0x17, 0x2c, 0x2e, 0xb9, 0x1a, 0xff, 0xd4, 0x19, 0xf1, 0x36, 0x78,
	0x9b, 0x3e, 0x6d, 0x00, 0x00, 0x00,
}

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
	out, err := ToJPEG(tinyWebP, DefaultQuality)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 3 || out[0] != 0xff || out[1] != 0xd8 {
		t.Fatalf("expected jpeg, got %d bytes", len(out))
	}
}

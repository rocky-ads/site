package imgconv

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/chai2010/webp"
)

const DefaultQuality = 80

func isJPEG(data []byte) bool {
	return len(data) >= 3 &&
		data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff
}

func isPNG(data []byte) bool {
	return bytes.HasPrefix(data, []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	})
}

func isWebP(data []byte) bool {
	return len(data) >= 12 &&
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WEBP"
}

// ToJPEG returns JPEG bytes. Existing JPEG is returned unchanged.
func ToJPEG(data []byte, quality int) ([]byte, error) {
	if isJPEG(data) {
		return data, nil
	}
	if quality <= 0 {
		quality = DefaultQuality
	}
	img, err := decode(data)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

func decode(data []byte) (image.Image, error) {
	r := bytes.NewReader(data)
	switch {
	case isPNG(data):
		img, err := png.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode png: %w", err)
		}
		return img, nil
	case isWebP(data):
		img, err := webp.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode webp: %w", err)
		}
		return img, nil
	default:
		img, _, err := image.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("decode image: %w", err)
		}
		return img, nil
	}
}

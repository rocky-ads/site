package handler

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/chai2010/webp"
	"mime/multipart"

	"github.com/chai2010/webp"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/logger"
	"golang.org/x/image/draw"
)

func uploadAdImages(store imagestore.Store, adID int,
	files []*multipart.FileHeader) {
	uploadAdImagesFromIndex(store, adID, 1, files)
}

func uploadAdImagesFromIndex(store imagestore.Store, adID, startIndex int,
	files []*multipart.FileHeader) {
	if len(files) == 0 {
		return
	}

	sizes := []struct {
		Width   int
		Suffix  string
		Quality float32
	}{
		{160, "160w", 60},
		{480, "480w", 70},
		{1200, "1200w", 80},
	}

	logger.Info("Starting ad image upload",
		"adID", adID, "imageCount", len(files), "startIndex", startIndex)

	for i, fileHeader := range files {
		imageIndex := startIndex + i
		file, err := fileHeader.Open()
		if err != nil {
			logger.Error("Failed to open uploaded image",
				"error", err, "adID", adID, "imageIndex", i+1)
			continue
		}

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(file); err != nil {
			logger.Error("Failed to read uploaded image",
				"error", err, "adID", adID, "imageIndex", i+1)
			file.Close()
			continue
		}
		file.Close()

		img, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
		if err != nil {
			logger.Error("Failed to decode uploaded image",
				"error", err, "adID", adID, "imageIndex", i+1)
			continue
		}

		bounds := img.Bounds()
		for _, sz := range sizes {
			w := sz.Width
			h := bounds.Dy() * w / bounds.Dx()
			dst := image.NewRGBA(image.Rect(0, 0, w, h))
			draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

			var webpBuf bytes.Buffer
			opt := &webp.Options{Lossless: false, Quality: sz.Quality}
			if err := webp.Encode(&webpBuf, dst, opt); err != nil {
				logger.Error("WebP encode error",
					"error", err, "adID", adID,
					"imageIndex", i+1, "size", sz.Suffix)
				continue
			}

			if err := store.Put(adID, imageIndex, sz.Suffix, webpBuf.Bytes()); err != nil {
				logger.Error("Image store put failed",
					"error", err, "adID", adID,
					"imageIndex", imageIndex, "size", sz.Suffix)
			}
		}
	}
}

package handler

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/rocky-ads/site/internal/config"
)

var allowedImageMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

func validateImageFile(fileHeader *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedImageExtensions[ext] {
		return fmt.Errorf(
			"file %s has invalid extension. Only .jpg, .jpeg, .png, "+
				".gif, and .webp are allowed",
			fileHeader.Filename,
		)
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		default:
			return fmt.Errorf(
				"unable to determine MIME type for file %s",
				fileHeader.Filename,
			)
		}
	}
	if !allowedImageMIMETypes[mimeType] {
		return fmt.Errorf(
			"file %s has invalid MIME type %s. Only image/jpeg, "+
				"image/png, image/gif, and image/webp are allowed",
			fileHeader.Filename, mimeType,
		)
	}

	if fileHeader.Size > config.ServerUploadLimit {
		return fmt.Errorf(
			"file %s exceeds maximum size of %d MB",
			fileHeader.Filename,
			config.ServerUploadLimit/(1024*1024),
		)
	}

	const minSize = int64(100)
	if fileHeader.Size < minSize {
		return fmt.Errorf(
			"file %s is too small and may be invalid",
			fileHeader.Filename,
		)
	}

	return nil
}

func validateImageFiles(files []*multipart.FileHeader) error {
	if len(files) > config.MaxImagesPerAd {
		return fmt.Errorf(
			"too many images. Maximum %d images allowed per ad",
			config.MaxImagesPerAd,
		)
	}
	for _, fileHeader := range files {
		if err := validateImageFile(fileHeader); err != nil {
			return err
		}
	}
	return nil
}

func validateAppendImageFiles(currentCount int,
	files []*multipart.FileHeader) error {
	if len(files) == 0 {
		return nil
	}
	if currentCount+len(files) > config.MaxImagesPerAd {
		return fmt.Errorf(
			"too many images. Maximum %d images allowed per ad",
			config.MaxImagesPerAd,
		)
	}
	for _, fileHeader := range files {
		if err := validateImageFile(fileHeader); err != nil {
			return err
		}
	}
	return nil
}

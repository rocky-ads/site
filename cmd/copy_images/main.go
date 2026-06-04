package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/chai2010/webp"
	"github.com/rocky-ads/site/internal/logger"
	"golang.org/x/image/draw"
)

// resizeImage resizes a webp image to the specified width (maintaining 4:3 aspect ratio)
func resizeImage(imageData []byte, targetWidth int) ([]byte, error) {
	// Decode webp image
	img, err := webp.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("decoding webp: %w", err)
	}

	// Calculate target height (4:3 aspect ratio)
	targetHeight := targetWidth * 3 / 4

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Resize using high-quality scaling
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)

	// Encode back to webp
	var buf bytes.Buffer
	if err := webp.Encode(&buf, dst, &webp.Options{Lossless: false, Quality: 85}); err != nil {
		return nil, fmt.Errorf("encoding webp: %w", err)
	}

	return buf.Bytes(), nil
}

func processImage(baseDir string, adID, imgIdx int) (created480, created160 bool, err error) {
	// Read 1200w image
	sourcePath := filepath.Join(baseDir, fmt.Sprintf("%d", adID), fmt.Sprintf("%d-1200w.webp", imgIdx))
	imageData, err := os.ReadFile(sourcePath)
	if err != nil {
		return false, false, fmt.Errorf("reading source image: %w", err)
	}

	// Check and create 480w if missing
	path480 := filepath.Join(baseDir, fmt.Sprintf("%d", adID), fmt.Sprintf("%d-480w.webp", imgIdx))
	if _, err := os.Stat(path480); os.IsNotExist(err) {
		logger.Info("Creating 480w image", "ad_id", adID, "img_idx", imgIdx)
		image480, err := resizeImage(imageData, 480)
		if err != nil {
			return false, false, fmt.Errorf("resizing to 480w: %w", err)
		}
		if err := os.WriteFile(path480, image480, 0644); err != nil {
			return false, false, fmt.Errorf("saving 480w image: %w", err)
		}
		logger.Info("Created 480w image", "ad_id", adID, "img_idx", imgIdx)
		created480 = true
	} else {
		logger.Debug("480w image already exists", "ad_id", adID, "img_idx", imgIdx)
	}

	// Check and create 160w if missing
	path160 := filepath.Join(baseDir, fmt.Sprintf("%d", adID), fmt.Sprintf("%d-160w.webp", imgIdx))
	if _, err := os.Stat(path160); os.IsNotExist(err) {
		logger.Info("Creating 160w image", "ad_id", adID, "img_idx", imgIdx)
		image160, err := resizeImage(imageData, 160)
		if err != nil {
			return created480, false, fmt.Errorf("resizing to 160w: %w", err)
		}
		if err := os.WriteFile(path160, image160, 0644); err != nil {
			return created480, false, fmt.Errorf("saving 160w image: %w", err)
		}
		logger.Info("Created 160w image", "ad_id", adID, "img_idx", imgIdx)
		created160 = true
	} else {
		logger.Debug("160w image already exists", "ad_id", adID, "img_idx", imgIdx)
	}

	return created480, created160, nil
}

func main() {
	baseDir := flag.String("dir", "static/images/ad", "Base directory containing ad images")
	flag.Parse()

	// Initialize logger
	if err := logger.Init("info", "text", ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting image copy process", "base_dir", *baseDir)

	// Pattern to match 1200w.webp files
	pattern := regexp.MustCompile(`^(\d+)-1200w\.webp$`)

	// Track statistics
	totalProcessed := 0
	totalCreated := 0
	errors := 0

	// Walk through all ad directories
	err := filepath.Walk(*baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if this is a 1200w.webp file
		filename := info.Name()
		matches := pattern.FindStringSubmatch(filename)
		if len(matches) != 2 {
			// Not a matching file, skip it
			return nil
		}

		// Extract image index
		imgIdx, err := strconv.Atoi(matches[1])
		if err != nil {
			logger.Warn("Invalid image index in filename", "filename", filename)
			return nil
		}

		// Extract ad ID from directory path
		dir := filepath.Dir(path)
		adDir := filepath.Base(dir)
		adID, err := strconv.Atoi(adDir)
		if err != nil {
			logger.Warn("Invalid ad ID in directory", "dir", dir)
			return nil
		}

		totalProcessed++

		// Process this image
		created480, created160, err := processImage(*baseDir, adID, imgIdx)
		if err != nil {
			logger.Error("Failed to process image", "ad_id", adID, "img_idx", imgIdx, "error", err)
			errors++
			return nil // Continue processing other images
		}

		// Track created images
		if created480 {
			totalCreated++
		}
		if created160 {
			totalCreated++
		}

		return nil
	})

	if err != nil {
		logger.Fatal("Error walking directory", "error", err)
	}

	logger.Info("Image copy process complete",
		"total_processed", totalProcessed,
		"total_created", totalCreated,
		"errors", errors)
}

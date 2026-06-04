package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/currency"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/logger"
	"golang.org/x/image/draw"
)

type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type FalRequest struct {
	Prompt       string     `json:"prompt"`
	ImageSize    *ImageSize `json:"image_size,omitempty"`
	OutputFormat string     `json:"output_format,omitempty"`
}

type FalImage struct {
	URL         string `json:"url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ContentType string `json:"content_type"`
}

type FalResponse struct {
	Images []FalImage `json:"images"`
}

func formatPrice(price int, priceCurrency string) string {
	return currency.Format(price, priceCurrency)
}

func formatLocation(city, adminArea, country string) string {
	parts := []string{}
	if city != "" {
		parts = append(parts, city)
	}
	if adminArea != "" {
		parts = append(parts, adminArea)
	}
	if country != "" {
		parts = append(parts, country)
	}
	return strings.Join(parts, ", ")
}

func buildPrompt(ad ad.Ad, isHandwrittenNote bool) string {
	location := formatLocation(ad.City, ad.AdminArea, ad.Country)
	price := formatPrice(ad.Price, ad.PriceCurrency)

	if isHandwrittenNote {
		return fmt.Sprintf("A hand-written note on paper or cardboard about: %s. The note mentions the title '%s', description '%s', price %s, and location %s. Written in casual handwriting, like someone selling something on craigslist or facebook marketplace. Non-professional photo quality, taken with a phone camera, natural lighting, slightly messy background.",
			ad.Title, ad.Title, ad.Description, price, location)
	}

	// Non-professional photo prompts - typical of craigslist/facebook marketplace
	promptVariations := []string{
		fmt.Sprintf("A casual photo of %s. %s. Price: %s. Location: %s. Non-professional photo taken with a phone camera, natural lighting, slightly messy background, typical of craigslist or facebook marketplace listing photos.",
			ad.Title, ad.Description, price, location),
		fmt.Sprintf("A quick snapshot photo of %s. Description: %s. Selling for %s in %s. Taken with a smartphone, amateur photography, realistic home lighting, cluttered background, authentic marketplace listing style.",
			ad.Title, ad.Description, price, location),
		fmt.Sprintf("An amateur photo of %s. %s. Price %s, located in %s. Phone camera quality, natural indoor or outdoor lighting, unprofessional composition, typical of online marketplace listings.",
			ad.Title, ad.Description, price, location),
		fmt.Sprintf("A casual marketplace photo of %s. Details: %s. Asking %s, location %s. Taken with a phone, realistic lighting, messy or cluttered background, authentic craigslist/facebook marketplace style.",
			ad.Title, ad.Description, price, location),
	}

	// Use a simple hash of ad ID to pick a variation (for consistency)
	variationIndex := ad.ID % len(promptVariations)
	return promptVariations[variationIndex]
}

func generateImage(apiKey, prompt string) ([]byte, error) {
	customSize := &ImageSize{
		Width:  1200,
		Height: 900, // 4:3 aspect ratio
	}

	requestBody := FalRequest{
		Prompt:       prompt,
		ImageSize:    customSize,
		OutputFormat: "webp",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := "https://fal.run/fal-ai/z-image/turbo"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Key "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var falResp FalResponse
	if err := json.Unmarshal(body, &falResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if len(falResp.Images) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	image := falResp.Images[0]

	// Download the image
	imgResp, err := http.Get(image.URL)
	if err != nil {
		return nil, fmt.Errorf("downloading image: %w", err)
	}
	defer imgResp.Body.Close()

	imageData, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading image data: %w", err)
	}

	return imageData, nil
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func deleteExistingImages(adID int) error {
	dir := fmt.Sprintf("static/images/ad/%d", adID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, nothing to delete
		}
		return err
	}

	for _, entry := range entries {
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func imageExists(adID, imgIdx int) bool {
	// imgIdx is 1-based (1, 2, 3, ...)
	filename := fmt.Sprintf("static/images/ad/%d/%d-1200w.webp", adID, imgIdx)
	_, err := os.Stat(filename)
	return err == nil
}

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

func saveImage(adID, imgIdx int, imageData []byte) error {
	// imgIdx is 1-based (1, 2, 3, ...)
	dir := fmt.Sprintf("static/images/ad/%d", adID)
	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Save 1200w image
	filename1200 := filepath.Join(dir, fmt.Sprintf("%d-1200w.webp", imgIdx))
	if err := os.WriteFile(filename1200, imageData, 0644); err != nil {
		return fmt.Errorf("saving 1200w image: %w", err)
	}

	// Create and save 480w version
	image480, err := resizeImage(imageData, 480)
	if err != nil {
		return fmt.Errorf("resizing to 480w: %w", err)
	}
	filename480 := filepath.Join(dir, fmt.Sprintf("%d-480w.webp", imgIdx))
	if err := os.WriteFile(filename480, image480, 0644); err != nil {
		return fmt.Errorf("saving 480w image: %w", err)
	}

	// Create and save 160w version
	image160, err := resizeImage(imageData, 160)
	if err != nil {
		return fmt.Errorf("resizing to 160w: %w", err)
	}
	filename160 := filepath.Join(dir, fmt.Sprintf("%d-160w.webp", imgIdx))
	if err := os.WriteFile(filename160, image160, 0644); err != nil {
		return fmt.Errorf("saving 160w image: %w", err)
	}

	return nil
}

func getAds(startID, limit int) ([]ad.Ad, error) {
	query := `
		SELECT
			a.id,
			a.category_id,
			a.title,
			a.description,
			a.price,
			a.price_currency,
			a.created_at,
			a.deleted_at,
			a.user_id,
			a.image_count,
			a.location_id,
			l.city,
			l.admin_area,
			l.country,
			0 AS bookmarked
		FROM ads a
		LEFT JOIN locations l ON a.location_id = l.id
		WHERE a.deleted_at IS NULL
	`
	args := []any{}
	if startID > 0 {
		query += " AND a.id >= ?"
		args = append(args, startID)
	}
	query += " ORDER BY a.id"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var ads []ad.Ad
	var err error
	if len(args) > 0 {
		err = db.Select(&ads, query, args...)
	} else {
		err = db.Select(&ads, query)
	}
	return ads, err
}

func main() {
	dbPath := flag.String("db", "project.db", "Path to database file")
	numAds := flag.Int("n", 0, "Number of ads to process (0 = all)")
	startID := flag.Int("s", 0, "Starting ad ID (process ads with ID >= this value)")
	noOverwrite := flag.Bool("x", false, "Don't overwrite existing files")
	flag.Parse()

	// Initialize logger
	if err := logger.Init("info", "text", ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Get API key
	apiKey := os.Getenv("FAL_API_KEY")
	if apiKey == "" {
		logger.Fatal("FAL_API_KEY environment variable is required")
	}

	// Initialize database
	if err := db.Init(*dbPath); err != nil {
		logger.Fatal("Failed to open database", "error", err)
	}
	defer db.Close()

	// Get ads
	logger.Info("Fetching ads from database...")
	ads, err := getAds(*startID, *numAds)
	if err != nil {
		logger.Fatal("Failed to fetch ads", "error", err)
	}

	if len(ads) == 0 {
		logger.Info("No ads found")
		return
	}

	logger.Info("Found ads", "count", len(ads))

	// Track total images generated for handwritten note logic
	totalImagesGenerated := 0

	// Process each ad
	for _, ad := range ads {
		logger.Info("Processing ad", "id", ad.ID, "title", ad.Title, "image_count", ad.ImageCount)

		// Delete existing images if -x is not set
		if !*noOverwrite {
			if err := deleteExistingImages(ad.ID); err != nil {
				logger.Error("Failed to delete existing images", "ad_id", ad.ID, "error", err)
				continue
			}
		}

		// Generate images for this ad (using 1-based indexing: 1, 2, 3, ...)
		for imgIdx := 1; imgIdx <= ad.ImageCount; imgIdx++ {
			// Check if image already exists and -x is set
			if *noOverwrite && imageExists(ad.ID, imgIdx) {
				logger.Info("Skipping existing image", "ad_id", ad.ID, "img_idx", imgIdx)
				totalImagesGenerated++
				continue
			}

			// Determine if this should be a handwritten note (every 30 images)
			isHandwrittenNote := (totalImagesGenerated+1)%30 == 0

			// Build prompt
			prompt := buildPrompt(ad, isHandwrittenNote)

			logger.Info("Generating image", "ad_id", ad.ID, "img_idx", imgIdx, "handwritten_note", isHandwrittenNote)

			// Generate image
			imageData, err := generateImage(apiKey, prompt)
			if err != nil {
				logger.Error("Failed to generate image", "ad_id", ad.ID, "img_idx", imgIdx, "error", err)
				continue
			}

			// Save image (imgIdx is already 1-based)
			if err := saveImage(ad.ID, imgIdx, imageData); err != nil {
				logger.Error("Failed to save image", "ad_id", ad.ID, "img_idx", imgIdx, "error", err)
				continue
			}

			logger.Info("Saved images", "ad_id", ad.ID, "img_idx", imgIdx, "sizes", "1200w, 480w, 160w")
			totalImagesGenerated++

			// Small delay to avoid rate limiting
			time.Sleep(500 * time.Millisecond)
		}
	}

	logger.Info("Image generation complete", "total_images", totalImagesGenerated)
}

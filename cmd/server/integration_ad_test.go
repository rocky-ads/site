package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
)

func TestCreateAdWithMileage(t *testing.T) {
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)

	clearCookie := &http.Cookie{Name: "auth_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: false}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{clearCookie})

	resp, _ := postFormRequest(t, baseURL+"/api/login", map[string]interface{}{
		"username": "test",
		"password": "test",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed with status %d", resp.StatusCode)
	}

	categoryCookie := &http.Cookie{Name: "category", Value: "6", Path: "/", HttpOnly: true, Secure: false}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	formData := map[string]interface{}{
		"title":          "Test Car Ad",
		"description":    "A test vehicle listing.",
		"year":           "2020",
		"price":          "15000",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
		"location":       "Los Angeles",
	}

	resp, result := postFormRequest(t, baseURL+"/auth/ad/new", formData)
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		t.Fatalf("expected redirect or success, got %d", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK {
		raw, _ := result["raw"].(string)
		if !strings.Contains(raw, "Test Car Ad") {
			t.Fatalf("expected created ad page, body=%q", raw)
		}
		return
	}
	loc := resp.Header.Get("Location")
	if !strings.HasPrefix(loc, "/ad/") {
		t.Fatalf("expected redirect to ad page, got %q", loc)
	}
}

func TestCreateAdWithImages(t *testing.T) {
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)
	loginTestUser(t, client, baseURLParsed)

	categoryCookie := &http.Cookie{
		Name: "category", Value: "6", Path: "/",
		HttpOnly: true, Secure: false,
	}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	pngData := minimalTestPNG(t)
	fields := map[string]string{
		"title":          "Ad With Images",
		"description":    "Listing with uploaded photos.",
		"year":           "2020",
		"price":          "15000",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
		"location":       "Los Angeles",
	}
	uploads := []multipartUpload{
		{fieldName: "images", fileName: "one.png", content: pngData},
		{fieldName: "images", fileName: "two.png", content: pngData},
	}

	resp, result := postMultipartRequest(t, baseURL+"/auth/ad/new", fields, uploads)
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		raw, _ := result["raw"].(string)
		t.Fatalf("expected redirect or success, got %d body=%q", resp.StatusCode, raw)
	}
	if resp.StatusCode == http.StatusOK {
		raw, _ := result["raw"].(string)
		if strings.Contains(raw, "error") || strings.Contains(raw, "required") {
			t.Fatalf("create failed: %q", raw)
		}
	}

	adID := adIDFromCreateResponse(t, resp, "Ad With Images")
	var imageCount int
	if err := db.QueryRow(
		"SELECT image_count FROM ads WHERE id = $1", adID,
	).Scan(&imageCount); err != nil {
		t.Fatal(err)
	}
	if imageCount != 2 {
		t.Fatalf("expected image_count 2, got %d", imageCount)
	}

	for _, size := range []string{"160w", "480w", "1200w"} {
		for idx := 1; idx <= 2; idx++ {
			path := filepath.Join(
				testImageDir, adID,
				fmt.Sprintf("%d-%s.webp", idx, size),
			)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("missing image file %s: %v", path, err)
			}
		}
	}
}

func TestCreateAdImageValidation(t *testing.T) {
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)
	loginTestUser(t, client, baseURLParsed)

	categoryCookie := &http.Cookie{
		Name: "category", Value: "6", Path: "/",
		HttpOnly: true, Secure: false,
	}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	baseFields := map[string]string{
		"title":          "Bad Image Ad",
		"description":    "Should not be created.",
		"year":           "2020",
		"price":          "15000",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
	}

	t.Run("invalid extension", func(t *testing.T) {
		resp, result := postMultipartRequest(t, baseURL+"/auth/ad/new", baseFields,
			[]multipartUpload{
				{fieldName: "images", fileName: "notes.txt", content: []byte("not an image")},
			})
		if resp.StatusCode == http.StatusFound {
			t.Fatal("expected validation failure, got redirect")
		}
		raw, _ := result["raw"].(string)
		if !strings.Contains(raw, "invalid extension") {
			t.Fatalf("expected extension error, body=%q", raw)
		}
	})

	t.Run("too many images", func(t *testing.T) {
		pngData := minimalTestPNG(t)
		uploads := make([]multipartUpload, config.MaxImagesPerAd+1)
		for i := range uploads {
			uploads[i] = multipartUpload{
				fieldName: "images",
				fileName:  fmt.Sprintf("img%d.png", i),
				content:   pngData,
			}
		}
		resp, result := postMultipartRequest(t, baseURL+"/auth/ad/new", baseFields, uploads)
		if resp.StatusCode == http.StatusFound {
			t.Fatal("expected validation failure, got redirect")
		}
		raw, _ := result["raw"].(string)
		if !strings.Contains(raw, "too many images") {
			t.Fatalf("expected count error, body=%q", raw)
		}
	})
}

func TestUpdateAd(t *testing.T) {
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)

	clearCookie := &http.Cookie{Name: "auth_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: false}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{clearCookie})

	resp, _ := postFormRequest(t, baseURL+"/api/login", map[string]interface{}{
		"username": "test",
		"password": "test",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed with status %d", resp.StatusCode)
	}

	categoryCookie := &http.Cookie{Name: "category", Value: "6", Path: "/", HttpOnly: true, Secure: false}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	createData := map[string]interface{}{
		"title":          "Before edit",
		"description":    "Original listing text.",
		"year":           "2020",
		"price":          "3400",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
	}
	resp, _ = postFormRequest(t, baseURL+"/auth/ad/new", createData)
	adID := adIDFromCreateResponse(t, resp, "Before edit")

	editData := map[string]interface{}{
		"title":                "After edit",
		"description_addition": "Has the 2.0L engine.",
		"price":                "3000",
		"price_currency":       "USD",
		"year":                 "2020",
		"mileage":              "12000",
		"mileage_unit":         "mi",
	}
	resp, _ = postFormRequest(t, baseURL+"/auth/ad/"+adID+"/edit", editData)
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		t.Fatalf("update ad: expected redirect or success, got %d", resp.StatusCode)
	}

	var desc string
	if err := db.QueryRow(
		"SELECT description FROM ads WHERE id = $1", adID,
	).Scan(&desc); err != nil {
		t.Fatal(err)
	}
	display := ad.DisplayDescription(desc)
	for _, want := range []string{
		"Original listing text.",
		"Description Addition",
		"Has the 2.0L engine.",
		"Title change",
		"Price change",
		"Price dropped",
	} {
		if !strings.Contains(display, want) {
			t.Errorf("description missing %q: %q", want, display)
		}
	}
}

func TestUpdateAdAppendImages(t *testing.T) {
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)
	loginTestUser(t, client, baseURLParsed)

	categoryCookie := &http.Cookie{
		Name: "category", Value: "6", Path: "/",
		HttpOnly: true, Secure: false,
	}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	pngData := minimalTestPNG(t)
	createFields := map[string]string{
		"title":          "Edit Append Images",
		"description":    "Original listing text.",
		"year":           "2020",
		"price":          "3400",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
	}
	createUploads := []multipartUpload{
		{fieldName: "images", fileName: "one.png", content: pngData},
	}
	resp, _ := postMultipartRequest(
		t, baseURL+"/auth/ad/new", createFields, createUploads,
	)
	adID := adIDFromCreateResponse(t, resp, "Edit Append Images")

	editFields := map[string]string{
		"title":          "Edit Append Images",
		"description":    "Original listing text.",
		"year":           "2020",
		"price":          "3400",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
	}
	editUploads := []multipartUpload{
		{fieldName: "images", fileName: "two.png", content: pngData},
	}
	resp, result := postMultipartRequest(
		t, baseURL+"/auth/ad/"+adID+"/edit", editFields, editUploads,
	)
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		raw, _ := result["raw"].(string)
		t.Fatalf("update ad: expected redirect or success, got %d body=%q",
			resp.StatusCode, raw)
	}

	var imageCount int
	var desc string
	if err := db.QueryRow(
		"SELECT image_count, description FROM ads WHERE id = $1", adID,
	).Scan(&imageCount, &desc); err != nil {
		t.Fatal(err)
	}
	if imageCount != 2 {
		t.Fatalf("expected image_count 2, got %d", imageCount)
	}
	display := ad.DisplayDescription(desc)
	if !strings.Contains(display, "Images Added") {
		t.Errorf("description missing Images Added: %q", display)
	}
	parts := ad.ParseDescriptionForDisplay(desc)
	if len(parts.History) == 0 ||
		len(parts.History[0].ImageIndices) != 1 ||
		parts.History[0].ImageIndices[0] != 2 {
		t.Fatalf("history indices = %v", parts.History[0].ImageIndices)
	}

	path := filepath.Join(
		testImageDir, adID, "2-160w.webp",
	)
	if _, err := os.Stat(path); err != nil {
		t.Errorf("missing appended image file %s: %v", path, err)
	}
}

func TestAdTagsPersistViaMultipartForm(t *testing.T) {
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)
	loginTestUser(t, client, baseURLParsed)

	categoryCookie := &http.Cookie{Name: "category", Value: "6", Path: "/", HttpOnly: true, Secure: false}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	createData := map[string]interface{}{
		"title":          "Tagged Ad",
		"description":    "A test vehicle listing.",
		"year":           "2020",
		"price":          "15000",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
		"location":       "Los Angeles",
		"suggestion": []string{
			ad.EncodeSuggestionFormValue("fuel", "diesel"),
			ad.EncodeSuggestionFormValue("clean title", "yes"),
		},
	}

	resp, _ := postFormRequest(t, baseURL+"/auth/ad/new", createData)
	adID := adIDFromCreateResponse(t, resp, "Tagged Ad")

	if tags := adTagsJSON(t, adID); tags == "[]" || tags == "" {
		t.Fatalf("tags not saved on create, got %q", tags)
	}

	editData := map[string]interface{}{
		"title":          "Tagged Ad",
		"price":          "15000",
		"price_currency": "USD",
		"year":           "2020",
		"mileage":        "12000",
		"mileage_unit":   "mi",
		"suggestion": []string{
			ad.EncodeSuggestionFormValue("fuel", "diesel"),
		},
	}
	resp, _ = postFormRequest(t, baseURL+"/auth/ad/"+adID+"/edit", editData)
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		t.Fatalf("update ad: expected redirect or success, got %d", resp.StatusCode)
	}
	if tags := adTagsJSON(t, adID); tags == "[]" || tags == "" {
		t.Fatalf("tags not preserved on edit, got %q", tags)
	}
}

func TestIntegrationGetAdNullLocationID(t *testing.T) {
	var adID int
	err := db.QueryRow(`INSERT INTO ads (category_id, title, description, user_id, location_id)
		VALUES ($1, 'No location ad', 'desc', $2, NULL) RETURNING id`,
		integrationCarsCategory, integrationTestUserID).Scan(&adID)
	if err != nil {
		t.Fatal(err)
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")
	got, err := ad.GetAd(0, adID, loc)
	if err != nil {
		t.Fatalf("GetAd(%d) failed: %v", adID, err)
	}
	if got.LocationID != nil {
		t.Fatalf("expected nil location_id, got %v", *got.LocationID)
	}
}

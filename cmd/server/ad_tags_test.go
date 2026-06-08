package main

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/db"
)

// TestAdTagsPersistViaMultipartForm guards against the multipart form regression:
// the ad form is submitted as multipart/form-data (it carries image uploads), so
// selected tag checkboxes must be parsed from the multipart body, not just from
// urlencoded PostArgs.
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

func adTagsJSON(t *testing.T, adID string) string {
	t.Helper()
	var tags string
	if err := db.QueryRow("SELECT tags FROM ads WHERE id = ?", adID).Scan(&tags); err != nil {
		t.Fatal(err)
	}
	return tags
}

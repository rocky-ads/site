package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/rocky-ads/site/internal/config"
)

func TestAdShareCanonicalURLAndFlyer(t *testing.T) {
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)
	loginTestUser(t, client, baseURLParsed)

	categoryCookie := &http.Cookie{
		Name: "category", Value: "6", Path: "/",
		HttpOnly: true, Secure: false,
	}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	formData := map[string]interface{}{
		"title":          "Flyer Share Ad",
		"description":    "Listing used to test share and flyer.",
		"year":           "2020",
		"price":          "15000",
		"price_currency": "USD",
		"mileage":        "12000",
		"mileage_unit":   "mi",
		"location":       "Los Angeles",
	}
	resp, _ := postFormRequest(t, baseURL+"/auth/ad/new", formData)
	adID := adIDFromCreateResponse(t, resp, "Flyer Share Ad")
	canonical := config.CanonicalURL("/ad/" + adID)
	flyerPath := fmt.Sprintf("/ad/%s/flyer", adID)

	resp, body := getRequestWithCookies(t, client,
		fmt.Sprintf("%s/ad/%s", baseURL, adID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ad detail status %d", resp.StatusCode)
	}
	if strings.Contains(body, flyerPath) {
		t.Fatal("ad detail should not include flyer link")
	}

	resp, body = getRequestWithCookies(t, client,
		fmt.Sprintf("%s/api/ad/%s/share", baseURL, adID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("share modal status %d", resp.StatusCode)
	}
	if !strings.Contains(body, canonical) {
		t.Fatalf("share modal missing CanonicalURL %q", canonical)
	}
	if !strings.Contains(body, flyerPath) {
		t.Fatal("share modal missing flyer link")
	}
	if !strings.Contains(body, "Print flyer") {
		t.Fatal("share modal missing print flyer button")
	}
	if strings.Contains(body, "flyer-thumb") {
		t.Fatal("share modal should not include flyer thumbnail")
	}

	resp, body = getRequestWithCookies(t, client, baseURL+flyerPath)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("owner flyer status %d", resp.StatusCode)
	}
	for _, want := range []string{
		"Flyer Share Ad",
		canonical,
		"data:image/png;base64,",
		"Print",
		"flyer-screen-only",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("flyer missing %q", want)
		}
	}
	if strings.Contains(body, `id="main-nav"`) {
		t.Fatal("flyer should not include site navigation")
	}

	resp, err := http.Get(baseURL + flyerPath)
	if err != nil {
		t.Fatal(err)
	}
	guestBody := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("guest flyer status %d", resp.StatusCode)
	}
	if !strings.Contains(guestBody, "Flyer Share Ad") {
		t.Fatal("guest flyer missing title")
	}
}

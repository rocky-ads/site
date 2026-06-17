package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/facet"
)

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestChromeDevToolsEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/.well-known/appspecific/com.chrome.devtools.json")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHomeHandler(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     int
		expectedStatus int
	}{
		{"Valid category", 6, 200},
		{"Another valid category", 5, 200},
		{"Invalid category", 999, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := getClientWithCategoryCookie(tt.categoryID)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			resp, body := getRequestWithCookies(t, client, baseURL+"/")

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				if len(body) > 500 {
					t.Logf("Response body (truncated): %s...", body[:500])
				} else {
					t.Logf("Response body: %s", body)
				}
			}

			if tt.expectedStatus == 200 {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/html" && contentType != "text/html; charset=utf-8" {
					t.Errorf("Expected Content-Type text/html, got %s", contentType)
				}
				if len(body) == 0 {
					t.Error("Expected non-empty HTML response")
				}

				if tt.categoryID == 999 {
					if !strings.Contains(body, "Car &amp; Truck Parts") && !strings.Contains(body, "Car & Truck Parts") {
						t.Errorf("Expected HTML to contain default category name 'Car & Truck Parts' (or HTML entity version)")
					}
				}
			}
		})
	}

	t.Run("No category cookie", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 without category cookie (should default to default category), got %d", resp.StatusCode)
		}

		hasCategoryCookie := false
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "category" {
				hasCategoryCookie = true
				break
			}
		}
		if !hasCategoryCookie {
			t.Error("Expected category cookie to be set when accessing home page without cookie")
		}
	})
}

func TestHomeHandlerFiltersExpanded(t *testing.T) {
	client, err := getClientWithCategoryCookie(6)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	if err := setSearchCookieOnClient(client, cookie.SearchState{
		Q: "Honda", Expanded: true,
	}); err != nil {
		t.Fatalf("set search cookie: %v", err)
	}
	resp, body := getRequestWithCookies(t, client, baseURL+"/")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "filter-price-min") {
		t.Error("Expected expanded filter panel with price field")
	}
	if !strings.Contains(body, `id="search-area"`) {
		t.Error("Expected #search-area on home page")
	}
	if !strings.Contains(body, `id="search-location"`) {
		t.Error("Expected #search-location under search box when filters expanded")
	}
	if !strings.Contains(body, `id="filter-toggle"`) {
		t.Error("Expected filter toggle on home page")
	}
	if !strings.Contains(body, `title="Collapse filters"`) {
		t.Error("Expected expanded filter toggle (^) when search cookie expanded")
	}
}

func TestSwitchCategoryHandler(t *testing.T) {
	tests := []struct {
		name                   string
		categoryID             string
		returnParam            string
		queryParams            string
		expectedStatus         int
		expectRedirect         string
		expectRedirectContains []string
	}{
		{"Valid category", "6", "", "", 200, "/", nil},
		{"Valid category with return", "5", "/auth/ad/new", "", 200, "/auth/ad/new", nil},
		{"Invalid category ID defaults to default category", "999", "", "", 200, "/", nil},
		{
			name:           "Redirect URL has no filter query params",
			categoryID:     "6",
			queryParams:    "q=Honda&price_min=10000",
			expectedStatus: 200,
			expectRedirect: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := getTestClient()

			requestURL := fmt.Sprintf("%s/api/category/%s/switch", baseURL, tt.categoryID)
			q := url.Values{}
			if tt.returnParam != "" {
				q.Set("return", tt.returnParam)
			}
			if tt.queryParams != "" {
				for _, part := range strings.Split(tt.queryParams, "&") {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) == 2 {
						q.Set(kv[0], kv[1])
					}
				}
			}
			if encoded := q.Encode(); encoded != "" {
				requestURL += "?" + encoded
			}
			resp, _ := getRequestWithCookies(t, client, requestURL)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				redirect := resp.Header.Get("HX-Redirect")
				if tt.expectRedirect != "" && redirect != tt.expectRedirect {
					t.Errorf("Expected HX-Redirect %q, got %q", tt.expectRedirect, redirect)
				}
				for _, s := range tt.expectRedirectContains {
					if !strings.Contains(redirect, s) {
						t.Errorf("Expected HX-Redirect to contain %q, got %q", s, redirect)
					}
				}

				hasCategoryCookie := false
				var cookieValue string
				for _, c := range resp.Cookies() {
					if c.Name == "category" {
						hasCategoryCookie = true
						cookieValue = c.Value
						break
					}
				}
				if !hasCategoryCookie {
					t.Error("Expected category cookie to be set")
				}
				expectedCookie := tt.categoryID
				if tt.categoryID == "999" {
					expectedCookie = "5"
				}
				if cookieValue != expectedCookie {
					t.Errorf("Expected category cookie value %s, got %s", expectedCookie, cookieValue)
				}
			}
		})
	}
}

func TestSwitchCategoryPreservesSearchCookie(t *testing.T) {
	client, err := getClientWithCategoryCookie(6)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	switchURL := baseURL + "/api/category/5/switch?q=Honda&price_min=10000"
	resp, _ := getRequestWithCookies(t, client, switchURL)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("switch: expected 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("HX-Redirect"); got != "/" {
		t.Fatalf("expected redirect /, got %q", got)
	}

	resp, body := getRequestWithCookies(t, client, baseURL+"/")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("home: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Honda") {
		t.Error("expected home page to show search query from cookie")
	}
}

func TestHideFiltersHandler(t *testing.T) {
	client, err := getClientWithCategoryCookie(6)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	min := 10000
	if err := setSearchCookieOnClient(client, cookie.SearchState{
		Q: "Honda", Facets: map[string]facet.Filter{"price": {Min: &min}}, Expanded: true,
	}); err != nil {
		t.Fatalf("set search cookie: %v", err)
	}

	resp, body := getRequestWithCookies(t, client, baseURL+"/api/show-filters")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("show-filters: expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(body, "filter-price-min") {
		t.Error("show-filters should render filter panel fragment")
	}
	if !strings.Contains(body, `id="filter-toggle"`) {
		t.Error("show-filters should OOB-swap filter toggle")
	}
	if !strings.Contains(body, `title="Collapse filters"`) {
		t.Error("show-filters should swap toggle to collapse (^)")
	}

	hideURL := baseURL + "/api/hide-filters?q=Honda&price_min=10000"
	resp, body = getRequestWithCookies(t, client, hideURL)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("hide-filters: expected 200, got %d", resp.StatusCode)
	}
	if strings.Contains(body, "filter-price-min") {
		t.Error("hide-filters should clear filter panel")
	}
	if !strings.Contains(body, `title="Expand filters"`) {
		t.Error("hide-filters should swap toggle to expand (v)")
	}
	if !strings.Contains(body, `id="search-results"`) {
		t.Error("hide-filters should OOB-swap search results")
	}
}

func TestCategorySelectHandler(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     int
		expectedStatus int
	}{
		{"Valid category", 6, 200},
		{"Another valid category", 5, 200},
		{"Invalid category", 999, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := getClientWithCategoryCookie(tt.categoryID)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			resp, body := getRequestWithCookies(t, client, baseURL+"/api/category-select")

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				if len(body) > 500 {
					t.Logf("Response body (truncated): %s...", body[:500])
				} else {
					t.Logf("Response body: %s", body)
				}
			}

			if tt.expectedStatus == 200 {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/html" && contentType != "text/html; charset=utf-8" {
					t.Errorf("Expected Content-Type text/html, got %s", contentType)
				}
				if len(body) == 0 {
					t.Error("Expected non-empty HTML response")
				}

				if tt.categoryID == 999 {
					if !strings.Contains(body, "Car &amp; Truck Parts") && !strings.Contains(body, "Car & Truck Parts") {
						t.Errorf("Expected HTML to contain default category name 'Car & Truck Parts' (or HTML entity version)")
					}
				}
			}
		})
	}

	t.Run("No category cookie", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/category-select")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 without category cookie (should default to default category), got %d", resp.StatusCode)
		}

		hasCategoryCookie := false
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "category" {
				hasCategoryCookie = true
				break
			}
		}
		if !hasCategoryCookie {
			t.Error("Expected category cookie to be set when accessing category select without cookie")
		}
	})
}

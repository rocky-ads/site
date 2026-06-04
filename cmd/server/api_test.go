package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/cmd/rebuild_db/seed"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/user"
)

var baseURL = "http://localhost:" + config.TestPort
var testServer *fiber.App
var testDBPath = "test.db"

// initDatabaseWithSchema initializes the database and loads the schema
func initDatabaseWithSchema(dbPath string) error {
	// Open database connection
	if err := db.Init(dbPath); err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	// Read and execute schema
	schemaPath := "internal/db/schema.sql"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Try from current working directory
		cwd, _ := os.Getwd()
		schemaPath = cwd + "/internal/db/schema.sql"
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("executing schema: %w", err)
	}

	return nil
}

// chdirModuleRoot sets the working directory to the module root (directory
// containing go.mod). Tests run with cwd in cmd/server; paths and seed data
// expect the repo root.
func chdirModuleRoot() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return os.Chdir(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// TestMain starts the test server before running tests and shuts it down after
func TestMain(m *testing.M) {
	if err := chdirModuleRoot(); err != nil {
		panic(fmt.Sprintf("Failed to chdir to module root: %v", err))
	}

	// Set test encryption keys before anything else
	// These must match the keys used when seeding the test database
	testUserEncryptionKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" // base64 encoded 32 bytes of zeros
	testMessageEncryptionKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	testJWTSecret := "test-jwt-secret-key-for-ci-minimum-32-chars-long"

	os.Setenv("USER_ENCRYPTION_KEY", testUserEncryptionKey)
	os.Setenv("MESSAGE_ENCRYPTION_KEY", testMessageEncryptionKey)
	os.Setenv("JWT_SECRET", testJWTSecret)

	// Update config variables using reflection since they're initialized at package import time
	// Decode and set UserEncryptionKey
	if key, err := base64.StdEncoding.DecodeString(testUserEncryptionKey); err == nil {
		configValue := reflect.ValueOf(&config.UserEncryptionKey).Elem()
		configValue.Set(reflect.ValueOf(key))
	}

	// Decode and set MessageEncryptionKey
	if key, err := base64.StdEncoding.DecodeString(testMessageEncryptionKey); err == nil {
		configValue := reflect.ValueOf(&config.MessageEncryptionKey).Elem()
		configValue.Set(reflect.ValueOf(key))
	}

	// Set JWTSecret
	configValue := reflect.ValueOf(&config.JWTSecret).Elem()
	configValue.Set(reflect.ValueOf([]byte(testJWTSecret)))

	reflect.ValueOf(&config.CookieSecure).Elem().SetBool(false)

	// Initialize logger for tests (use minimal logging)
	if err := logger.Init("error", "text", ""); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Remove existing test database if it exists
	if _, err := os.Stat(testDBPath); err == nil {
		if err := os.Remove(testDBPath); err != nil {
			panic(fmt.Sprintf("Failed to remove existing test database: %v", err))
		}
	}

	// Initialize database with schema
	if err := initDatabaseWithSchema(testDBPath); err != nil {
		panic(fmt.Sprintf("Failed to initialize test database: %v", err))
	}

	// Load seed data (including test ads)
	if err := seed.LoadAll(true); err != nil {
		db.Close()
		os.Remove(testDBPath)
		panic(fmt.Sprintf("Failed to load seed data: %v", err))
	}

	// Database is already initialized and seeded, now load categories
	if err := ad.LoadCategories(); err != nil {
		db.Close()
		os.Remove(testDBPath)
		panic(fmt.Sprintf("Failed to initialize ads: %v", err))
	}

	// Setup test server
	testServer = setupApp()

	// Start server in a goroutine
	port := ":" + config.TestPort
	go func() {
		if err := testServer.Listen(port); err != nil {
			panic(fmt.Sprintf("Test server failed to start: %v", err))
		}
	}()

	// Wait for server to be ready
	maxAttempts := 30
	for i := range maxAttempts {
		resp, err := http.Get("http://localhost" + port + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		if i == maxAttempts-1 {
			panic("Test server failed to start within timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Run tests
	code := m.Run()

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := testServer.ShutdownWithContext(ctx); err != nil {
		fmt.Printf("Error shutting down test server: %v\n", err)
	}

	// Close database
	db.Close()

	// Clean up test database
	if _, err := os.Stat(testDBPath); err == nil {
		if err := os.Remove(testDBPath); err != nil {
			fmt.Printf("Warning: Failed to remove test database: %v\n", err)
		}
	}

	os.Exit(code)
}

// Shared HTTP client with persistent cookie jar for CSRF token caching
var testClient *http.Client
var testClientOnce sync.Once

func getTestClient() *http.Client {
	testClientOnce.Do(func() {
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(fmt.Sprintf("Failed to create cookie jar: %v", err))
		}
		testClient = &http.Client{
			Jar: jar,
		}
		// Initialize CSRF token by making a GET request to /health
		baseURLParsed, _ := url.Parse(baseURL)
		getReq, _ := http.NewRequest("GET", baseURL+"/health", nil)
		getResp, err := testClient.Do(getReq)
		if err == nil {
			getResp.Body.Close()
			// Handle Secure cookie issue for HTTP testing
			for _, cookie := range getResp.Cookies() {
				if cookie.Name == "_csrf" && cookie.Secure {
					testCookie := &http.Cookie{
						Name:     cookie.Name,
						Value:    cookie.Value,
						Path:     cookie.Path,
						Domain:   cookie.Domain,
						HttpOnly: cookie.HttpOnly,
						SameSite: cookie.SameSite,
						Secure:   false,
					}
					jar.SetCookies(baseURLParsed, []*http.Cookie{testCookie})
				}
			}
		}
	})
	return testClient
}

func postFormRequest(t *testing.T, requestURL string, body map[string]interface{}) (*http.Response, map[string]interface{}) {
	client := getTestClient()

	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("Failed to parse base URL: %v", err)
	}

	var csrfToken string
	cookies := client.Jar.Cookies(baseURLParsed)
	for _, cookie := range cookies {
		if cookie.Name == "_csrf" {
			csrfToken = cookie.Value
			break
		}
	}

	if csrfToken == "" {
		getReq, err := http.NewRequest("GET", baseURL+"/health", nil)
		if err != nil {
			t.Fatalf("Failed to create GET request for CSRF token: %v", err)
		}
		getResp, err := client.Do(getReq)
		if err != nil {
			t.Fatalf("Failed to get CSRF token: %v", err)
		}
		getResp.Body.Close()

		for _, cookie := range getResp.Cookies() {
			if cookie.Name == "_csrf" {
				csrfToken = cookie.Value
				testCookie := &http.Cookie{
					Name:     cookie.Name,
					Value:    cookie.Value,
					Path:     cookie.Path,
					Domain:   cookie.Domain,
					HttpOnly: cookie.HttpOnly,
					SameSite: cookie.SameSite,
					Secure:   false,
				}
				client.Jar.SetCookies(baseURLParsed, []*http.Cookie{testCookie})
				break
			}
		}
	}

	if csrfToken == "" {
		t.Fatalf("Failed to get CSRF token from cookie")
	}

	var bodyBuffer bytes.Buffer
	writer := multipart.NewWriter(&bodyBuffer)

	for k, v := range body {
		switch val := v.(type) {
		case []string:
			for _, s := range val {
				fieldWriter, err := writer.CreateFormField(k)
				if err != nil {
					t.Fatalf("Failed to create form field %s: %v", k, err)
				}
				fieldWriter.Write([]byte(s))
			}
		case string:
			fieldWriter, err := writer.CreateFormField(k)
			if err != nil {
				t.Fatalf("Failed to create form field %s: %v", k, err)
			}
			fieldWriter.Write([]byte(val))
		}
	}

	writer.Close()

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		t.Fatalf("Failed to parse request URL: %v", err)
	}

	req, err := http.NewRequest("POST", requestURL, &bodyBuffer)
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Csrf-Token", csrfToken)

	jarCookies := client.Jar.Cookies(reqURL)
	for _, cookie := range jarCookies {
		if cookie.Name == "_csrf" {
			req.AddCookie(cookie)
			break
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	bodyRespBytes, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(bodyRespBytes, &result); err != nil {
		return resp, map[string]interface{}{"raw": string(bodyRespBytes)}
	}
	return resp, result
}

// Test Health Check
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

// Test GET /api/search/ hard filters
func TestSearchPageHandler(t *testing.T) {
	tests := []struct {
		name           string
		categoryID     int
		query          string
		expectContains []string
		expectAbsent   []string
	}{
		{
			name:           "text query matches title",
			categoryID:     6,
			query:          "?q=Honda",
			expectContains: []string{"search-results"},
		},
		{
			name:           "price min filter",
			categoryID:     6,
			query:          "?price_min=10000",
			expectContains: []string{"search-results"},
		},
		{
			name:           "location and radius",
			categoryID:     6,
			query:          "?location=Denver&radius=50",
			expectContains: []string{"search-results"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := getClientWithCategoryCookie(tt.categoryID)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			if tt.query != "?q=Honda" {
				if err := setSearchCookieOnClient(client, cookie.SearchState{Expanded: true}); err != nil {
					t.Fatalf("set search cookie: %v", err)
				}
			}
			url := baseURL + "/api/search/" + tt.query
			resp, body := getRequestWithCookies(t, client, url)
			if resp.StatusCode != http.StatusOK {
				snippet := body
				if len(snippet) > 200 {
					snippet = snippet[:200]
				}
				t.Fatalf("Expected status 200, got %d body=%s", resp.StatusCode, snippet)
			}
			for _, s := range tt.expectContains {
				if !strings.Contains(body, s) {
					t.Errorf("expected body to contain %q", s)
				}
			}
			for _, s := range tt.expectAbsent {
				if strings.Contains(body, s) {
					t.Errorf("expected body not to contain %q", s)
				}
			}
		})
	}
}

// Helper function to create a client with category cookie set
func getClientWithCategoryCookie(categoryID int) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	// Get CSRF token via GET /health
	getReq, err := http.NewRequest("GET", baseURL+"/health", nil)
	if err != nil {
		return nil, err
	}
	getResp, err := client.Do(getReq)
	if err != nil {
		return nil, err
	}
	getResp.Body.Close()

	// Set category cookie
	categoryCookie := &http.Cookie{
		Name:     "category",
		Value:    strconv.Itoa(categoryID),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

	// Handle CSRF cookie (copy from response if needed)
	for _, cookie := range getResp.Cookies() {
		if cookie.Name == "_csrf" {
			testCookie := &http.Cookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Domain:   cookie.Domain,
				HttpOnly: cookie.HttpOnly,
				SameSite: cookie.SameSite,
				Secure:   false,
			}
			jar.SetCookies(baseURLParsed, []*http.Cookie{testCookie})
			break
		}
	}

	return client, nil
}

func setSearchCookieOnClient(client *http.Client, state cookie.SearchState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{{
		Name:  "search",
		Value: base64.RawURLEncoding.EncodeToString(data),
		Path:  "/",
	}})
	return nil
}

// Helper function to make GET request with cookies
func getRequestWithCookies(t *testing.T, client *http.Client, url string) (*http.Response, string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}

	bodyBytes := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyBytes = append(bodyBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	resp.Body.Close()

	return resp, string(bodyBytes)
}

// Test GET /.well-known/appspecific/com.chrome.devtools.json
func TestChromeDevToolsEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/.well-known/appspecific/com.chrome.devtools.json")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Static file serving returns 200 OK (not 204 No Content)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// Test GET /
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
				// Check that it returns HTML (Fiber sets Content-Type as "text/html" without charset)
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/html" && contentType != "text/html; charset=utf-8" {
					t.Errorf("Expected Content-Type text/html, got %s", contentType)
				}
				// Check that body contains some expected HTML elements
				if len(body) == 0 {
					t.Error("Expected non-empty HTML response")
				}

				// If invalid category (999), verify results are for default category
				if tt.categoryID == 999 {
					// Verify HTML contains default category name (check for HTML entity version)
					if !strings.Contains(body, "Car &amp; Truck Parts") && !strings.Contains(body, "Car & Truck Parts") {
						t.Errorf("Expected HTML to contain default category name 'Car & Truck Parts' (or HTML entity version)")
					}
				}
			}
		})
	}

	// Test without category cookie - should default to default category
	t.Run("No category cookie", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 without category cookie (should default to default category), got %d", resp.StatusCode)
		}

		// Verify that category cookie was set
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
	if !strings.Contains(body, `id="filter-panel"`) {
		t.Error("Expected #filter-panel on home page")
	}
	if !strings.Contains(body, `id="search-bar"`) {
		t.Error("Expected #search-bar on home page")
	}
	if !strings.Contains(body, `id="filter-toggle"`) {
		t.Error("Expected filter toggle on home page")
	}
	if !strings.Contains(body, `title="Collapse filters"`) {
		t.Error("Expected expanded filter toggle (^) when search cookie expanded")
	}
}

// Test GET /login
func TestLoginHandler(t *testing.T) {
	resp, err := http.Get(baseURL + "/login")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that it returns HTML (Fiber sets Content-Type as "text/html" without charset)
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html" && contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}

	// Read body to verify it's HTML
	bodyBytes := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyBytes = append(bodyBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	body := string(bodyBytes)
	if len(body) == 0 {
		t.Error("Expected non-empty HTML response")
	}
}

// Test POST /api/login
func TestLoginSubmitHandler(t *testing.T) {
	// Check if test user exists - fail test if prerequisites not met
	// Also check if encryption key is configured (needed for decryption)
	if len(config.UserEncryptionKey) == 0 {
		t.Fatal("USER_ENCRYPTION_KEY environment variable not set. This is required for user decryption.")
	}

	testUser, err := user.GetByName("test")
	if err != nil {
		// Check if it's a "not found" error vs decryption error
		if errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("Test user not found in database. Database may need to be seeded with: ./rebuild_db -test-ads")
		} else {
			t.Fatalf("Failed to retrieve test user (error: %v). This might be a decryption error if USER_ENCRYPTION_KEY doesn't match the key used to seed the database.", err)
		}
	}

	// Verify we got a valid user
	if testUser.ID == 0 {
		t.Fatal("Test user found but invalid (ID is 0)")
	}

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectJWT      bool
	}{
		{"Valid login", "test", "test", 200, true},
		{"Valid login - admin", "admin", "admin", 200, true},
		{"Invalid username", "nonexistent", "test", 200, false},
		{"Invalid password", "test", "wrong", 200, false},
		{"Empty username", "", "test", 200, false},
		{"Empty password", "test", "", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := getTestClient()
			baseURLParsed, _ := url.Parse(baseURL)

			// Clear any existing auth_token cookie from previous tests
			// Create an expired cookie to clear it
			clearCookie := &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteStrictMode,
			}
			client.Jar.SetCookies(baseURLParsed, []*http.Cookie{clearCookie})

			// Create form data
			formData := map[string]interface{}{
				"username": tt.username,
				"password": tt.password,
			}

			resp, result := postFormRequest(t, baseURL+"/api/login", formData)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				if len(result) > 0 {
					t.Logf("Response: %+v", result)
				}
			}

			// Check for JWT cookie ONLY in response cookies (not jar, to avoid cookies from previous tests)
			hasJWT := false
			var jwtCookie *http.Cookie
			for _, cookie := range resp.Cookies() {
				if cookie.Name == "auth_token" {
					hasJWT = true
					jwtCookie = cookie
					// Handle Secure cookie for HTTP testing (add to jar if login succeeded)
					if cookie.Secure && tt.expectJWT {
						testCookie := &http.Cookie{
							Name:     cookie.Name,
							Value:    cookie.Value,
							Path:     cookie.Path,
							Domain:   cookie.Domain,
							HttpOnly: cookie.HttpOnly,
							SameSite: cookie.SameSite,
							Secure:   false,
						}
						client.Jar.SetCookies(baseURLParsed, []*http.Cookie{testCookie})
					}
					break
				}
			}

			if tt.expectJWT {
				if !hasJWT {
					t.Error("Expected JWT cookie but it was not set")
					// Debug: log all cookies
					t.Logf("Response cookies: %v", resp.Cookies())
				} else if jwtCookie != nil && jwtCookie.Value == "" {
					t.Error("JWT cookie value is empty")
				}
			} else {
				if hasJWT {
					t.Error("Unexpected JWT cookie for failed login")
				}
			}
		})
	}
}

// Test GET /api/category/:category/switch
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
				// ParseCategory maps invalid IDs (e.g. 999) to default category (5)
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

// Test GET /api/hide-filters clears the filter panel and OOB-refreshes results.
func TestHideFiltersHandler(t *testing.T) {
	client, err := getClientWithCategoryCookie(6)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	min := 10000
	if err := setSearchCookieOnClient(client, cookie.SearchState{
		Q: "Honda", PriceMin: &min, Expanded: true,
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

// Test GET /api/category-select
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
				// Check that it returns HTML (Fiber sets Content-Type as "text/html" without charset)
				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/html" && contentType != "text/html; charset=utf-8" {
					t.Errorf("Expected Content-Type text/html, got %s", contentType)
				}
				// Check that body contains some expected HTML elements
				if len(body) == 0 {
					t.Error("Expected non-empty HTML response")
				}

				// If invalid category (999), verify results are for default category
				if tt.categoryID == 999 {
					// Verify HTML contains default category name (check for HTML entity version)
					if !strings.Contains(body, "Car &amp; Truck Parts") && !strings.Contains(body, "Car & Truck Parts") {
						t.Errorf("Expected HTML to contain default category name 'Car & Truck Parts' (or HTML entity version)")
					}
				}
			}
		})
	}

	// Test without category cookie - should default to default category
	t.Run("No category cookie", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/category-select")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 without category cookie (should default to default category), got %d", resp.StatusCode)
		}

		// Verify that category cookie was set
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

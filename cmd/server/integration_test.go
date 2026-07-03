package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
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
	"github.com/rocky-ads/site/cmd/seed_db/seed"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/db/testdb"
	"github.com/rocky-ads/site/internal/handler"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/vector"
)

var baseURL = "http://localhost:" + config.TestPort
var testServer *fiber.App
var testImageDir string

const (
	integrationTestUserID     = 1
	integrationInquirerUserID = 2
	integrationCarsCategory   = 6
	integrationGarageCategory = 7
)

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

	if err := testdb.InitSchema(); err != nil {
		panic(fmt.Sprintf("Failed to initialize test database: %v", err))
	}

	if err := seed.LoadAll(); err != nil {
		db.Close()
		panic(fmt.Sprintf("Failed to load seed data: %v", err))
	}

	// Database is already initialized and seeded, now load categories
	if err := ad.LoadCategories(); err != nil {
		db.Close()
		panic(fmt.Sprintf("Failed to initialize ads: %v", err))
	}

	vector.SetEmbedder(vector.NewFakeEmbedder())
	if err := vector.InitEmbeddingCaches(); err != nil {
		db.Close()
		panic(fmt.Sprintf("Failed to init embedding caches: %v", err))
	}
	if _, err := vector.BackfillAllAdsSync(); err != nil {
		db.Close()
		panic(fmt.Sprintf("Failed to backfill ad embeddings: %v", err))
	}
	vector.StartBackgroundProcessor()

	var err error
	testImageDir, err = os.MkdirTemp("", "test-ad-images-*")
	if err != nil {
		db.Close()
		panic(fmt.Sprintf("Failed to create test image dir: %v", err))
	}
	handler.SetAdImageStore(imagestore.NewLocal(testImageDir))

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

	if testImageDir != "" {
		os.RemoveAll(testImageDir)
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

func postFormRequest(t *testing.T, requestURL string,
	body map[string]interface{}) (*http.Response, map[string]interface{}) {
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
		req.AddCookie(cookie)
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

type multipartUpload struct {
	fieldName string
	fileName  string
	content   []byte
}

func postMultipartRequest(t *testing.T, requestURL string,
	fields map[string]string, uploads []multipartUpload) (*http.Response, map[string]interface{}) {
	t.Helper()
	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)

	var csrfToken string
	for _, cookie := range client.Jar.Cookies(baseURLParsed) {
		if cookie.Name == "_csrf" {
			csrfToken = cookie.Value
			break
		}
	}
	if csrfToken == "" {
		getReq, err := http.NewRequest("GET", baseURL+"/health", nil)
		if err != nil {
			t.Fatalf("create GET request for CSRF token: %v", err)
		}
		getResp, err := client.Do(getReq)
		if err != nil {
			t.Fatalf("get CSRF token: %v", err)
		}
		getResp.Body.Close()
		for _, cookie := range getResp.Cookies() {
			if cookie.Name == "_csrf" {
				csrfToken = cookie.Value
				client.Jar.SetCookies(baseURLParsed, []*http.Cookie{
					{
						Name: cookie.Name, Value: cookie.Value,
						Path: cookie.Path, Domain: cookie.Domain,
						HttpOnly: cookie.HttpOnly, SameSite: cookie.SameSite,
						Secure: false,
					},
				})
				break
			}
		}
	}
	if csrfToken == "" {
		t.Fatal("Failed to get CSRF token from cookie")
	}

	var bodyBuffer bytes.Buffer
	writer := multipart.NewWriter(&bodyBuffer)
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatalf("write field %s: %v", k, err)
		}
	}
	for _, u := range uploads {
		part, err := writer.CreateFormFile(u.fieldName, u.fileName)
		if err != nil {
			t.Fatalf("create form file %s: %v", u.fileName, err)
		}
		if _, err := part.Write(u.content); err != nil {
			t.Fatalf("write form file %s: %v", u.fileName, err)
		}
	}
	writer.Close()

	reqURL, err := url.Parse(requestURL)
	if err != nil {
		t.Fatalf("parse request URL: %v", err)
	}

	req, err := http.NewRequest("POST", requestURL, &bodyBuffer)
	if err != nil {
		t.Fatalf("create POST request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Csrf-Token", csrfToken)
	for _, cookie := range client.Jar.Cookies(reqURL) {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST request: %v", err)
	}
	defer resp.Body.Close()

	bodyRespBytes, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(bodyRespBytes, &result); err != nil {
		return resp, map[string]interface{}{"raw": string(bodyRespBytes)}
	}
	return resp, result
}

func minimalTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func loginTestUser(t *testing.T, client *http.Client, baseURLParsed *url.URL) {
	t.Helper()
	clearCookie := &http.Cookie{
		Name: "auth_token", Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: false,
	}
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{clearCookie})
	resp, _ := postFormRequest(t, baseURL+"/api/login", map[string]interface{}{
		"username": "test",
		"password": "test",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed with status %d", resp.StatusCode)
	}
}

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

	getReq, err := http.NewRequest("GET", baseURL+"/health", nil)
	if err != nil {
		return nil, err
	}
	getResp, err := client.Do(getReq)
	if err != nil {
		return nil, err
	}
	getResp.Body.Close()

	categoryCookie := &http.Cookie{
		Name:     "category",
		Value:    strconv.Itoa(categoryID),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}
	jar.SetCookies(baseURLParsed, []*http.Cookie{categoryCookie})

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

func setSearchCookieOnClient(client *http.Client,
	state cookie.SearchState) error {
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

func getSearchStateFromClient(client *http.Client) (cookie.SearchState, error) {
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return cookie.SearchState{}, err
	}
	for _, c := range client.Jar.Cookies(baseURLParsed) {
		if c.Name != "search" {
			continue
		}
		data, err := base64.RawURLEncoding.DecodeString(c.Value)
		if err != nil {
			return cookie.SearchState{}, err
		}
		var state cookie.SearchState
		if err := json.Unmarshal(data, &state); err != nil {
			return cookie.SearchState{}, err
		}
		return state, nil
	}
	return cookie.SearchState{}, nil
}

func getRequestWithCookies(t *testing.T,
	client *http.Client, url string) (*http.Response, string) {
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

func adIDFromCreateResponse(t *testing.T, resp *http.Response,
	title string) string {
	t.Helper()
	if resp.StatusCode == http.StatusFound {
		loc := resp.Header.Get("Location")
		if strings.HasPrefix(loc, "/ad/") {
			return strings.TrimPrefix(loc, "/ad/")
		}
	}
	var id int
	err := db.QueryRow(
		"SELECT id FROM ads WHERE title = $1 ORDER BY id DESC LIMIT 1", title,
	).Scan(&id)
	if err != nil {
		t.Fatalf("create ad: status %d, lookup id: %v", resp.StatusCode, err)
	}
	return strconv.Itoa(id)
}

func adTagsJSON(t *testing.T, adID string) string {
	t.Helper()
	var tags string
	if err := db.QueryRow("SELECT tags FROM ads WHERE id = $1", adID).Scan(&tags); err != nil {
		t.Fatal(err)
	}
	return tags
}

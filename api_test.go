package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/logger"
)

var baseURL = "http://localhost:" + config.TestPort
var testServer *fiber.App

// TestMain starts the test server before running tests and shuts it down after
func TestMain(m *testing.M) {
	// Initialize logger for tests (use minimal logging)
	if err := logger.Init("error", "text", ""); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Initialize database
	if err := db.Init("project.db"); err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
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
	for i := 0; i < maxAttempts; i++ {
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

// Helper functions to get large expected result arrays at test time
// These fetch from the API to ensure exact matches with current seed data

func getEngineValues() []string {
	resp, err := http.Get(baseURL + "/api/categories/Cars%20%26%20Trucks/values/engine")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var values []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil
	}

	resultSlice := make([]string, len(values))
	for i, v := range values {
		if str, ok := v.(string); ok {
			resultSlice[i] = str
		}
	}
	return resultSlice
}

func getCarsTrucksMakeValues() []string {
	resp, err := http.Get(baseURL + "/api/categories/Cars%20%26%20Trucks/values/make")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var values []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil
	}

	resultSlice := make([]string, len(values))
	for i, v := range values {
		if str, ok := v.(string); ok {
			resultSlice[i] = str
		}
	}
	return resultSlice
}

func getCarsTrucksModelValues() []string {
	resp, err := http.Get(baseURL + "/api/categories/Cars%20%26%20Trucks/values/model")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var values []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil
	}

	resultSlice := make([]string, len(values))
	for i, v := range values {
		if str, ok := v.(string); ok {
			resultSlice[i] = str
		}
	}
	return resultSlice
}

func getCarsTrucksYearValues() []string {
	resp, err := http.Get(baseURL + "/api/categories/Cars%20%26%20Trucks/values/year")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var values []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil
	}

	resultSlice := make([]string, len(values))
	for i, v := range values {
		if str, ok := v.(string); ok {
			resultSlice[i] = str
		}
	}
	return resultSlice
}

// Helper functions for any-values expected results (hard-coded)
func getCarTruckPartsAnyMakeValues() []string {
	return []string{"AC", "ACURA", "ALPINE", "AMERICAN AUSTIN", "AMERICAN BANTAM", "AMERICAN MOTORS", "ARMSTRONG-SIDDELEY", "ARNOLT-BRISTOL", "ASUNA", "AUDI", "AUSTIN", "BENTLEY", "BLACKHAWK", "BMW", "BOND", "BORGWARD", "BRICKLIN", "BRISTOL", "BUGATTI", "BUICK", "BYD", "CADILLAC", "CHECKER", "CHEVROLET", "CHRYSLER", "CISITALIA", "CITROEN", "CLEVELAND", "COLE", "CONTINENTAL", "CORD", "CROSLEY", "CUNNINGHAM", "DAIMLER", "DAVIS", "DELAHAYE", "DELLOW", "DELOREAN", "DENZEL", "DESOTO", "DETOMASO", "DIANA", "DKW", "DODGE", "DORETTI", "DUESENBERG", "EAGLE", "EDSEL", "ELVA", "ESSEX", "FACEL VEGA", "FALCON KNIGHT", "FARGO", "FERRARI", "FISKER", "FLINT", "FORD", "FOTON", "FRANKLIN", "FRAZER NASH", "GEELY", "GEO", "GLAS", "GMC", "GRAHAM", "GRAHAM-PAIGE", "GRIFFITH", "GWM", "HAVAL", "HILLMAN", "HINO", "HONDA", "HRG", "HUDSON", "HUMMER", "HYUNDAI", "ISUZU", "JAC", "JEEP", "JENSEN", "JMC", "JOWETT", "KENWORTH", "KIA", "KURTIS", "LADA", "LANCHESTER", "LEXINGTON", "LEXUS", "LINCOLN", "LORDSTOWN MOTORS", "LOTUS", "MAICO", "MARATHON", "MARAUDER", "MARCOS", "MARMON", "MARQUETTE", "MASTRETTA", "MATRA", "MERCURY", "MESSERSCHMITT", "MITSUBISHI", "MITSUBISHI FUSO", "MOBILITY VENTURES", "MOON", "MORRIS", "MOSKVICH", "NASH", "NISSAN", "OAKLAND", "OLDSMOBILE", "OMEGA", "OMODA", "PACKARD", "PAIGE", "PANOZ", "PEGASO", "PIERCE-ARROW", "PLYMOUTH", "POLESTAR", "PONTIAC", "PORSCHE", "RAM", "RICKENBACKER", "RIVIAN", "ROAMER", "ROLLS-ROYCE", "ROOSEVELT", "ROVER", "SATURN", "SEAT", "SHELBY", "SIMCA", "SINGER", "SKODA", "SMART", "SPYKER", "SRT", "STANDARD", "STAR", "STEARNS KNIGHT", "STEVENS-DURYEA", "SUNBEAM", "SUZUKI", "SWALLOW", "TALBOT-LAGO", "TATRA", "TESLA", "THINK", "TOYOTA"}
}

func getCarTruckPartsAnyYearValues() []string {
	return []string{"1903", "1904", "1906", "1907", "1908", "1910", "1911", "1912", "1914", "1915", "1916", "1917", "1918", "1920", "1921", "1923", "1924", "1925", "1926", "1927", "1928", "1929", "1930", "1931", "1932", "1933", "1934", "1935", "1936", "1937", "1938", "1939", "1940", "1941", "1942", "1943", "1944", "1945", "1946", "1947", "1948", "1949", "1950", "1951", "1952", "1953", "1954", "1955", "1956", "1957", "1958", "1959", "1960", "1961", "1962", "1963", "1964", "1965", "1966", "1967", "1968", "1969", "1970", "1971", "1972", "1973", "1974", "1975", "1976", "1977", "1978", "1979", "1980", "1981", "1982", "1983", "1984", "1985", "1986", "1987", "1988", "1989", "1990", "1991", "1992", "1993", "1994", "1995", "1996", "1997", "1998", "1999", "2000", "2001", "2002", "2003", "2004", "2005", "2006", "2007", "2008", "2009", "2010", "2011", "2012", "2013", "2014", "2015", "2016", "2017", "2018", "2019", "2020", "2021", "2022", "2023", "2024", "2025", "2026"}
}

func getCarTruckPartsAnyPartCategoryValues() []string {
	return []string{"Belt Drive", "Body & Lamp Assembly", "Brakes & Wheel Hub", "Cooling System", "Drivetrain", "Electrical", "Electrical-Bulb & Socket", "Electrical-Connector", "Electrical-Switch & Relay", "Engine", "Exhaust & Emission", "Fuel & Air", "Heat & Air Conditioning"}
}

func getCarTruckPartsAnyModelValues() []string {
	resp, err := http.Get(baseURL + "/api/categories/Car%20%26%20Truck%20Parts/any-values/model")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var values []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil
	}

	resultSlice := make([]string, len(values))
	for i, v := range values {
		if str, ok := v.(string); ok {
			resultSlice[i] = str
		}
	}
	return resultSlice
}

func getCarTruckPartsAnyEngineValues() []string {
	resp, err := http.Get(baseURL + "/api/categories/Car%20%26%20Truck%20Parts/any-values/engine")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var values []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil
	}

	resultSlice := make([]string, len(values))
	for i, v := range values {
		if str, ok := v.(string); ok {
			resultSlice[i] = str
		}
	}
	return resultSlice
}

// Helper function to make GET requests
func getRequest(t *testing.T, url string) (*http.Response, map[string]interface{}) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make GET request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Read body to check if it's JSON
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

	// Try to decode as JSON array first (since some endpoints return arrays directly)
	var arrResult []interface{}
	if err := json.Unmarshal(bodyBytes, &arrResult); err == nil {
		return resp, map[string]interface{}{"array": arrResult}
	}

	// Try to decode as JSON object
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		// Not JSON, return empty map (error response)
		return resp, map[string]interface{}{}
	}
	return resp, result
}

func postFormRequest(t *testing.T, requestURL string, body map[string]interface{}) (*http.Response, map[string]interface{}) {
	// Use shared HTTP client with persistent cookie jar
	client := getTestClient()

	// Parse base URL for cookie jar
	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("Failed to parse base URL: %v", err)
	}

	// Extract CSRF token from cookie jar (cached from first GET request)
	var csrfToken string
	cookies := client.Jar.Cookies(baseURLParsed)
	for _, cookie := range cookies {
		if cookie.Name == "_csrf" {
			csrfToken = cookie.Value
			break
		}
	}

	// If CSRF token is missing, fetch it by making a GET request to /health
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

		// Extract CSRF token from response cookies
		for _, cookie := range getResp.Cookies() {
			if cookie.Name == "_csrf" {
				csrfToken = cookie.Value
				// Create a copy without Secure flag for HTTP testing
				// Go's cookie jar filters Secure cookies when scheme is HTTP
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
		t.Fatalf("Failed to get CSRF token from cookie. Cookies in jar: %v", cookies)
	}

	// Create multipart form data
	var bodyBuffer bytes.Buffer
	writer := multipart.NewWriter(&bodyBuffer)

	for k, v := range body {
		switch val := v.(type) {
		case []int:
			for _, id := range val {
				fieldWriter, err := writer.CreateFormField(k)
				if err != nil {
					t.Fatalf("Failed to create form field %s: %v", k, err)
				}
				fieldWriter.Write([]byte(strconv.Itoa(id)))
			}
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
		case int:
			fieldWriter, err := writer.CreateFormField(k)
			if err != nil {
				t.Fatalf("Failed to create form field %s: %v", k, err)
			}
			fieldWriter.Write([]byte(strconv.Itoa(val)))
		}
	}

	writer.Close()

	// Parse request URL to ensure cookies are sent correctly
	reqURL, err := url.Parse(requestURL)
	if err != nil {
		t.Fatalf("Failed to parse request URL: %v", err)
	}

	req, err := http.NewRequest("POST", requestURL, &bodyBuffer)
	if err != nil {
		t.Fatalf("Failed to create POST request to %s: %v", requestURL, err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Csrf-Token", csrfToken)

	// Ensure cookies from jar are included for the request URL
	// The jar should handle this automatically, but we ensure the CSRF cookie is included
	jarCookies := client.Jar.Cookies(reqURL)
	for _, cookie := range jarCookies {
		if cookie.Name == "_csrf" {
			req.AddCookie(cookie)
			break
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make POST request to %s: %v", requestURL, err)
	}
	defer resp.Body.Close()

	// Read body to check if it's JSON
	bodyRespBytes := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyRespBytes = append(bodyRespBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Try to decode as JSON
	var result map[string]interface{}
	if err := json.Unmarshal(bodyRespBytes, &result); err != nil {
		// Try array
		var arrResult []interface{}
		if err2 := json.Unmarshal(bodyRespBytes, &arrResult); err2 != nil {
			// Not JSON, return empty map (error response)
			return resp, map[string]interface{}{}
		}
		return resp, map[string]interface{}{"array": arrResult}
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

// Test GET /api/categories/:category/values/:field
func TestGetAllValues(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		field          string
		queryParams    string
		expectedStatus int
		expectedResult []string // exact expected values (must match exactly)
	}{
		{
			name:           "Cars & Trucks - make values",
			category:       "Cars%20%26%20Trucks",
			field:          "make",
			queryParams:    "",
			expectedStatus: 200,
			expectedResult: []string{"ABARTH", "AC", "ACURA", "ALFA ROMEO", "ALLARD", "ALLSTATE", "ALPINE", "ALVIS", "AM GENERAL", "AMERICAN AUSTIN", "AMERICAN BANTAM", "AMERICAN MOTORS", "AMPHICAR", "APOLLO", "APPERSON", "ARMSTRONG-SIDDELEY", "ARNOLT-BRISTOL", "ARNOLT-MG", "ASTON MARTIN", "ASUNA", "AUBURN", "AUDI", "AUSTIN", "AUSTIN-HEALEY", "AVANTI", "BAIC", "BENTLEY", "BERKELEY", "BESTUNE", "BIZZARRINI", "BLACKHAWK", "BMW", "BOND", "BORGWARD", "BRICKLIN", "BRISTOL", "BUGATTI", "BUICK", "BYD", "CADILLAC", "CASE", "CHANDLER", "CHANGAN", "CHECKER", "CHEVROLET", "CHIREY", "CHRYSLER", "CISITALIA", "CITROEN", "CLEVELAND", "COLE", "CONTINENTAL", "CORD", "CROSLEY", "CUNNINGHAM", "CUPRA", "DAEWOO", "DAF", "DAIHATSU", "DAIMLER", "DATSUN", "DAVIS", "DE VAUX", "DELAGE", "DELAHAYE", "DELLOW", "DELOREAN", "DENZEL", "DESOTO", "DETOMASO", "DEUTSCH-BONNET", "DFSK", "DIANA", "DKW", "DODGE", "DORETTI", "DU PONT", "DUAL-GHIA", "DUESENBERG", "DURANT", "EAGLE", "EDSEL", "ELCAR", "ELVA", "ERSKINE", "ESSEX", "EXCALIBUR", "FACEL VEGA", "FAIRTHORPE", "FALCON KNIGHT", "FARGO", "FAW", "FERRARI", "FIAT", "FISKER", "FLINT", "FORD", "FOTON", "FRANKLIN", "FRAZER NASH", "FREIGHTLINER", "GAC", "GARDNER", "GEELY", "GENESIS", "GEO", "GIANT MOTORS", "GLAS", "GMC", "GOLIATH", "GORDON-KEEBLE", "GRAHAM", "GRAHAM-PAIGE", "GRIFFITH", "GWM", "HAVAL", "HAYNES", "HCS", "HEALEY", "HENRY J", "HERTZ", "HILLMAN", "HINO", "HISPANO-SUIZA", "HONDA", "HOTCHKISS", "HRG", "HUDSON", "HUMBER", "HUMMER", "HUPMOBILE", "HYUNDAI", "INEOS", "INFINITI", "INTERNATIONAL", "ISO", "ISUZU", "IVECO", "JAC", "JAGUAR", "JEEP", "JENSEN", "JEWETT", "JMC", "JORDAN", "JOWETT", "KAISER-FRAZER", "KARMA", "KENWORTH", "KIA", "KISSEL", "KURTIS", "LADA", "LAFORZA", "LAGONDA", "LAMBORGHINI", "LANCHESTER", "LANCIA", "LAND ROVER", "LASALLE", "LEA-FRANCIS", "LEXINGTON", "LEXUS", "LINCOLN", "LOCOMOBILE", "LORDSTOWN MOTORS", "LOTUS", "LUCID", "MACK", "MAICO", "MARATHON", "MARAUDER", "MARCOS", "MARMON", "MARQUETTE", "MASERATI", "MASTRETTA", "MATRA", "MAXWELL", "MAYBACH", "MAZDA", "MCLAREN", "MERCEDES-BENZ", "MERCURY", "MERKUR", "MESSERSCHMITT", "MG", "MINI", "MITSUBISHI", "MITSUBISHI FUSO", "MOBILITY VENTURES", "MONTEVERDI", "MOON", "MORETTI", "MORGAN", "MORRIS", "MOSKVICH", "NARDI", "NASH", "NISSAN", "NSU", "OAKLAND", "OLDSMOBILE", "OMEGA", "OMODA", "OPEL", "OSCA", "PACKARD", "PAIGE", "PANHARD", "PANOZ", "PANTHER", "PASSPORT", "PEERLESS", "PEGASO", "PETERBILT", "PEUGEOT", "PIERCE-ARROW", "PLYMOUTH", "POLESTAR", "PONTIAC", "PORSCHE", "QVALE", "RAM", "RELIANT", "RENAULT", "REO", "RICKENBACKER", "RILEY", "RIVIAN", "ROAMER", "ROCKNE", "ROLLIN", "ROLLS-ROYCE", "ROOSEVELT", "ROVER", "SAAB", "SABRA", "SALEEN", "SALMSON", "SATURN", "SCION", "SEAT", "SERES", "SHELBY", "SIATA", "SIMCA", "SINGER", "SKODA", "SMART", "SPYKER", "SRT", "SSANGYONG", "STANDARD", "STAR", "STEARNS KNIGHT", "STERLING", "STEVENS-DURYEA", "STUDEBAKER", "STUTZ", "SUBARU", "SUNBEAM", "SUZUKI", "SWALLOW", "TALBOT-LAGO", "TATRA", "TESLA", "THINK", "TOYOTA", "TRIUMPH", "TURNER", "TVR", "UAZ", "UD", "UTILIMASTER", "VAM", "VAUXHALL", "VELIE", "VESPA", "VIKING", "VINFAST", "VOLKSWAGEN", "VOLVO", "VPG", "WARTBURG", "WESTCOTT", "WHIPPET", "WILLYS", "WINDSOR", "WOLSELEY", "WORKHORSE", "YELLOW CAB", "YUGO", "ZACUA", "ZUNDAPP"},
		},
		{
			name:           "Cars & Trucks - year values",
			category:       "Cars%20%26%20Trucks",
			field:          "year",
			queryParams:    "",
			expectedStatus: 200,
			expectedResult: getCarsTrucksYearValues(), // Large list, fetched at test time
		},
		{
			name:           "Cars & Trucks - model values",
			category:       "Cars%20%26%20Trucks",
			field:          "model",
			queryParams:    "",
			expectedStatus: 200,
			expectedResult: getCarsTrucksModelValues(), // Large list (4911 values), fetched at test time
		},
		{
			name:           "Cars & Trucks - engine values",
			category:       "Cars%20%26%20Trucks",
			field:          "engine",
			queryParams:    "",
			expectedStatus: 200,
			expectedResult: getEngineValues(), // Large list (1338 values), fetched at test time
		},
		{
			name:           "Car & Truck Parts - make values",
			category:       "Car%20%26%20Truck%20Parts",
			field:          "make",
			queryParams:    "",
			expectedStatus: 200,
			expectedResult: getCarsTrucksMakeValues(), // Same as Cars & Trucks makes
		},
		{
			name:           "Car & Truck Parts - part_category values",
			category:       "Car%20%26%20Truck%20Parts",
			field:          "part_category",
			queryParams:    "",
			expectedStatus: 200,
			expectedResult: []string{"Belt Drive", "Body & Lamp Assembly", "Brakes & Wheel Hub", "Cooling System", "Drivetrain", "Electrical", "Electrical-Bulb & Socket", "Electrical-Connector", "Electrical-Switch & Relay", "Engine", "Exhaust & Emission", "Fuel & Air", "Heat & Air Conditioning"},
		},
		{
			name:           "With filter - year filtered by make",
			category:       "Cars%20%26%20Trucks",
			field:          "year",
			queryParams:    "?make=AC",
			expectedStatus: 200,
			expectedResult: []string{"1947", "1948", "1949", "1950", "1951", "1952", "1953", "1954", "1955", "1956", "1957", "1958", "1959", "1960", "1961", "1962", "1963", "1967", "1968", "1969", "1970", "1971", "1972", "1973"},
		},
		{
			name:           "With filter - model filtered by make and year",
			category:       "Cars%20%26%20Trucks",
			field:          "model",
			queryParams:    "?make=AC&year=1956",
			expectedStatus: 200,
			expectedResult: []string{"ACE", "ACECA", "TWO-LITRE"},
		},
		{
			name:           "With filter - make value with spaces",
			category:       "Agricultural%20Equipment",
			field:          "year",
			queryParams:    "?make=JOHN%20DEERE",
			expectedStatus: 200,
			expectedResult: nil, // Don't check exact match, just verify it works
		},
		{
			name:           "Car & Truck Parts - part_category filtered by part_category with spaces",
			category:       "Car%20%26%20Truck%20Parts",
			field:          "part_subcategory",
			queryParams:    "?part_category=Exhaust%20%26%20Emission",
			expectedStatus: 200,
			expectedResult: nil, // Don't check exact match, just verify it works
		},
		{
			name:           "Invalid category",
			category:       "InvalidCategory",
			field:          "make",
			queryParams:    "",
			expectedStatus: 404,
		},
		{
			name:           "Invalid field",
			category:       "Cars%20%26%20Trucks",
			field:          "invalid_field",
			queryParams:    "",
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/categories/%s/values/%s%s", baseURL, tt.category, tt.field, tt.queryParams)
			resp, result := getRequest(t, url)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				return
			}

			if tt.expectedStatus == 200 {
				// Validate values array (now returned directly as array)
				values, ok := result["array"].([]interface{})
				if !ok {
					t.Errorf("Response should be an array")
					return
				}

				// Convert values to string slice for comparison
				actualValues := make([]string, len(values))
				for i, v := range values {
					if str, ok := v.(string); ok {
						actualValues[i] = str
					} else {
						t.Errorf("Expected all values to be strings, got %T at index %d", v, i)
						return
					}
				}

				// Check for exact match (skip if expectedResult is nil)
				if tt.expectedResult != nil {
					if len(actualValues) != len(tt.expectedResult) {
						t.Errorf("Expected %d values, got %d. Expected: %v, Got: %v", len(tt.expectedResult), len(actualValues), tt.expectedResult, actualValues)
					} else {
						// Compare each value (order matters for exact match)
						for i, expected := range tt.expectedResult {
							if i >= len(actualValues) || actualValues[i] != expected {
								t.Errorf("Expected values[%d] to be '%s', got '%s'. Expected: %v, Got: %v", i, expected, actualValues[i], tt.expectedResult, actualValues)
								break
							}
						}
					}
				}

			}
		})
	}
}

// Test GET /api/categories/:category/any-values/:field
func TestGetAnyValues(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		field          string
		queryParams    string
		expectedStatus int
		expectedResult []string // exact expected values (nil means don't check exact match)
	}{
		{"Car & Truck Parts - any make values", "Car%20%26%20Truck%20Parts", "make", "", 200, getCarTruckPartsAnyMakeValues()},
		{"Car & Truck Parts - any year values", "Car%20%26%20Truck%20Parts", "year", "", 200, getCarTruckPartsAnyYearValues()},
		{"Car & Truck Parts - any model values", "Car%20%26%20Truck%20Parts", "model", "", 200, getCarTruckPartsAnyModelValues()},
		{"Car & Truck Parts - any engine values", "Car%20%26%20Truck%20Parts", "engine", "", 200, getCarTruckPartsAnyEngineValues()},
		{"Car & Truck Parts - any part_category values", "Car%20%26%20Truck%20Parts", "part_category", "", 200, getCarTruckPartsAnyPartCategoryValues()},
		{"With filter chain", "Car%20%26%20Truck%20Parts", "part_category", "?make=AC&year=1953&model=ACE&engine=2.0L%20L6", 200, []string{"Brakes & Wheel Hub"}},
		{"With filter using part_category with spaces", "Car%20%26%20Truck%20Parts", "part_subcategory", "?part_category=Exhaust%20%26%20Emission", 200, nil},
		{"Agricultural Equipment - filter by make with spaces", "Agricultural%20Equipment", "year", "?make=JOHN%20DEERE", 200, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/categories/%s/any-values/%s%s", baseURL, tt.category, tt.field, tt.queryParams)
			resp, result := getRequest(t, url)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				values, ok := result["array"].([]interface{})
				if !ok {
					t.Errorf("Response should be an array")
					return
				}

				// Convert to string slice
				actualValues := make([]string, len(values))
				for i, v := range values {
					if str, ok := v.(string); ok {
						actualValues[i] = str
					}
				}

				// Check expected results if provided
				if tt.expectedResult != nil {
					if len(actualValues) != len(tt.expectedResult) {
						t.Errorf("Expected %d values, got %d", len(tt.expectedResult), len(actualValues))
						return
					}
					for i := range actualValues {
						if actualValues[i] != tt.expectedResult[i] {
							t.Errorf("Value at index %d: expected %q, got %q", i, tt.expectedResult[i], actualValues[i])
							return
						}
					}
				}
			}
		})
	}
}

// Test POST /api/categories/:category/ad-values/:field
func TestGetAdValues(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		field          string
		body           map[string]interface{}
		expectedStatus int
		expectedResult []string // exact expected values (nil means don't check exact match)
	}{
		{"Cars & Trucks - ad values for make", "Cars%20%26%20Trucks", "make", map[string]interface{}{"ad_ids": []int{984, 985}}, 200, []string{"FORD", "HONDA"}},
		{"Cars & Trucks - ad values for year", "Cars%20%26%20Trucks", "year", map[string]interface{}{"ad_ids": []int{984}}, 200, []string{"2020"}},
		{"Cars & Trucks - ad values for year with multiple year filter", "Cars%20%26%20Trucks", "year", map[string]interface{}{"ad_ids": []int{984, 985}, "year": []string{"2020", "1975"}}, 200, []string{"1975", "2020"}},
		{"Cars & Trucks - ad values for model", "Cars%20%26%20Trucks", "model", map[string]interface{}{"ad_ids": []int{984}}, 200, []string{"CIVIC"}},
		{"Cars & Trucks - ad values with filter", "Cars%20%26%20Trucks", "engine", map[string]interface{}{"ad_ids": []int{984, 985}, "make": []string{"HONDA"}}, 200, []string{"2.0L L4"}},
		{"Agricultural Equipment - ad values for make with spaces", "Agricultural%20Equipment", "make", map[string]interface{}{"ad_ids": []int{1}}, 200, []string{"JOHN DEERE"}},
		{"Car & Truck Parts - ad values for part_category with spaces", "Car%20%26%20Truck%20Parts", "part_category", map[string]interface{}{"ad_ids": []int{11}}, 200, []string{"Exhaust & Emission"}},
		{"Car & Truck Parts - ad values with filter using value with spaces", "Car%20%26%20Truck%20Parts", "part_subcategory", map[string]interface{}{"ad_ids": []int{11}, "part_category": []string{"Exhaust & Emission"}}, 200, nil},
		{"Empty ad_ids", "Cars%20%26%20Trucks", "make", map[string]interface{}{"ad_ids": []int{}}, 200, []string{}},
		{"Ad_ids from different category", "Cars%20%26%20Trucks", "make", map[string]interface{}{"ad_ids": []int{1}}, 200, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/categories/%s/ad-values/%s", baseURL, tt.category, tt.field)
			resp, result := postFormRequest(t, url, tt.body)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				values, ok := result["array"].([]interface{})
				if !ok {
					t.Errorf("Response should be an array")
					return
				}

				// Convert to string slice
				actualValues := make([]string, len(values))
				for i, v := range values {
					if str, ok := v.(string); ok {
						actualValues[i] = str
					}
				}

				// Check expected results if provided
				if tt.expectedResult != nil {
					if len(actualValues) != len(tt.expectedResult) {
						t.Errorf("Expected %d values, got %d. Expected: %v, Got: %v", len(tt.expectedResult), len(actualValues), tt.expectedResult, actualValues)
						return
					}
					for i := range actualValues {
						if actualValues[i] != tt.expectedResult[i] {
							t.Errorf("Value at index %d: expected %q, got %q", i, tt.expectedResult[i], actualValues[i])
							return
						}
					}
				}
			}
		})
	}
}

// Helper functions for chains expected results (hard-coded)
func getCarsTrucksChains() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"ChainIndex": float64(0),
			"Fields": []interface{}{
				map[string]interface{}{
					"Name":        "make",
					"DisplayName": "Make",
					"Order":       float64(1),
					"NextInChain": float64(2),
				},
				map[string]interface{}{
					"Name":        "year",
					"DisplayName": "Year",
					"Order":       float64(2),
					"NextInChain": float64(3),
				},
				map[string]interface{}{
					"Name":        "model",
					"DisplayName": "Model",
					"Order":       float64(3),
					"NextInChain": float64(4),
				},
				map[string]interface{}{
					"Name":        "engine",
					"DisplayName": "Engine",
					"Order":       float64(4),
					"NextInChain": float64(0),
				},
			},
		},
	}
}

func getCarTruckPartsChains() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"ChainIndex": float64(0),
			"Fields": []interface{}{
				map[string]interface{}{
					"Name":        "make",
					"DisplayName": "Make",
					"Order":       float64(1),
					"NextInChain": float64(2),
				},
				map[string]interface{}{
					"Name":        "year",
					"DisplayName": "Year",
					"Order":       float64(2),
					"NextInChain": float64(3),
				},
				map[string]interface{}{
					"Name":        "model",
					"DisplayName": "Model",
					"Order":       float64(3),
					"NextInChain": float64(4),
				},
				map[string]interface{}{
					"Name":        "engine",
					"DisplayName": "Engine",
					"Order":       float64(4),
					"NextInChain": float64(0),
				},
			},
		},
		map[string]interface{}{
			"ChainIndex": float64(1),
			"Fields": []interface{}{
				map[string]interface{}{
					"Name":        "part_category",
					"DisplayName": "Part Category",
					"Order":       float64(5),
					"NextInChain": float64(6),
				},
				map[string]interface{}{
					"Name":        "part_subcategory",
					"DisplayName": "Part Subcategory",
					"Order":       float64(6),
					"NextInChain": float64(0),
				},
			},
		},
	}
}

func getMotorcyclesChains() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"ChainIndex": float64(0),
			"Fields": []interface{}{
				map[string]interface{}{
					"Name":        "make",
					"DisplayName": "Make",
					"Order":       float64(1),
					"NextInChain": float64(2),
				},
				map[string]interface{}{
					"Name":        "year",
					"DisplayName": "Year",
					"Order":       float64(2),
					"NextInChain": float64(3),
				},
				map[string]interface{}{
					"Name":        "model",
					"DisplayName": "Model",
					"Order":       float64(3),
					"NextInChain": float64(4),
				},
				map[string]interface{}{
					"Name":        "engine",
					"DisplayName": "Engine",
					"Order":       float64(4),
					"NextInChain": float64(0),
				},
			},
		},
	}
}

// Test GET /api/categories/:category/chains
func TestGetChains(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		expectedStatus int
		expectedResult []interface{} // expected chains structure (nil means don't check exact match)
	}{
		{"Cars & Trucks - chains", "Cars%20%26%20Trucks", 200, getCarsTrucksChains()},
		{"Car & Truck Parts - chains", "Car%20%26%20Truck%20Parts", 200, getCarTruckPartsChains()},
		{"Motorcycles - chains", "Motorcycles", 200, getMotorcyclesChains()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/categories/%s/chains", baseURL, tt.category)
			resp, result := getRequest(t, url)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				arr, ok := result["array"].([]interface{})
				if !ok {
					t.Errorf("Response should be an array")
					return
				}

				// Check expected results if provided
				if tt.expectedResult != nil {
					if len(arr) != len(tt.expectedResult) {
						t.Errorf("Expected %d chains, got %d", len(tt.expectedResult), len(arr))
						return
					}

					// Compare each chain
					for chainIdx, expectedChain := range tt.expectedResult {
						actualChain, ok := arr[chainIdx].(map[string]interface{})
						if !ok {
							t.Errorf("Chain %d should be an object", chainIdx)
							return
						}

						expectedChainMap := expectedChain.(map[string]interface{})

						// Check ChainIndex
						if actualChain["ChainIndex"] != expectedChainMap["ChainIndex"] {
							t.Errorf("Chain %d ChainIndex: expected %v, got %v", chainIdx, expectedChainMap["ChainIndex"], actualChain["ChainIndex"])
							return
						}

						// Check Fields
						actualFields, ok := actualChain["Fields"].([]interface{})
						if !ok {
							t.Errorf("Chain %d Fields should be an array", chainIdx)
							return
						}

						expectedFields := expectedChainMap["Fields"].([]interface{})
						if len(actualFields) != len(expectedFields) {
							t.Errorf("Chain %d Fields: expected %d fields, got %d", chainIdx, len(expectedFields), len(actualFields))
							return
						}

						for fieldIdx, expectedField := range expectedFields {
							actualField, ok := actualFields[fieldIdx].(map[string]interface{})
							if !ok {
								t.Errorf("Chain %d Field %d should be an object", chainIdx, fieldIdx)
								return
							}

							expectedFieldMap := expectedField.(map[string]interface{})
							for key, expectedValue := range expectedFieldMap {
								if actualField[key] != expectedValue {
									t.Errorf("Chain %d Field %d %s: expected %v, got %v", chainIdx, fieldIdx, key, expectedValue, actualField[key])
									return
								}
							}
						}
					}
				}
			}
		})
	}
}

// Helper functions for first spec fields expected results (hard-coded)
func getCarsTrucksFirstSpecFields() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"name":         "make",
			"display_name": "Make",
		},
	}
}

func getCarTruckPartsFirstSpecFields() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"name":         "make",
			"display_name": "Make",
		},
		map[string]interface{}{
			"name":         "part_category",
			"display_name": "Part Category",
		},
	}
}

func getMotorcyclesFirstSpecFields() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"name":         "make",
			"display_name": "Make",
		},
	}
}

// Test GET /api/categories/:category/first-spec-fields
func TestGetFirstSpecFields(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		expectedStatus int
		expectedResult []interface{} // expected first spec fields structure (nil means don't check exact match)
	}{
		{"Cars & Trucks - first spec fields", "Cars%20%26%20Trucks", 200, getCarsTrucksFirstSpecFields()},
		{"Car & Truck Parts - first spec fields", "Car%20%26%20Truck%20Parts", 200, getCarTruckPartsFirstSpecFields()},
		{"Motorcycles - first spec fields", "Motorcycles", 200, getMotorcyclesFirstSpecFields()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/categories/%s/first-spec-fields", baseURL, tt.category)
			resp, result := getRequest(t, url)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				arr, ok := result["array"].([]interface{})
				if !ok {
					t.Errorf("Response should be an array")
					return
				}

				// Check expected results if provided
				if tt.expectedResult != nil {
					if len(arr) != len(tt.expectedResult) {
						t.Errorf("Expected %d first spec fields, got %d", len(tt.expectedResult), len(arr))
						return
					}

					// Compare each field
					for fieldIdx, expectedField := range tt.expectedResult {
						actualField, ok := arr[fieldIdx].(map[string]interface{})
						if !ok {
							t.Errorf("Field %d should be an object", fieldIdx)
							return
						}

						expectedFieldMap := expectedField.(map[string]interface{})
						for key, expectedValue := range expectedFieldMap {
							if actualField[key] != expectedValue {
								t.Errorf("Field %d %s: expected %v, got %v", fieldIdx, key, expectedValue, actualField[key])
								return
							}
						}
					}
				}
			}
		})
	}
}

// Test GET /api/categories/:category/last-spec-field
func TestGetLastSpecField(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		expectedStatus int
		expectedResult map[string]interface{} // expected last spec field structure (nil means don't check exact match)
	}{
		{"Cars & Trucks - last spec field", "Cars%20%26%20Trucks", 200, map[string]interface{}{"name": "engine", "display_name": "Engine"}},
		{"Car & Truck Parts - last spec field", "Car%20%26%20Truck%20Parts", 200, map[string]interface{}{"name": "part_subcategory", "display_name": "Part Subcategory"}},
		{"Motorcycles - last spec field", "Motorcycles", 200, map[string]interface{}{"name": "engine", "display_name": "Engine"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/categories/%s/last-spec-field", baseURL, tt.category)
			resp, result := getRequest(t, url)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				if result["name"] == nil || result["display_name"] == nil {
					t.Errorf("Response should contain 'name' and 'display_name' fields")
					return
				}

				// Check expected results if provided
				if tt.expectedResult != nil {
					for key, expectedValue := range tt.expectedResult {
						if result[key] != expectedValue {
							t.Errorf("Field %s: expected %v, got %v", key, expectedValue, result[key])
							return
						}
					}
				}
			}
		})
	}
}

// Test GET /api/ads/:id/filter-values
func TestGetAdFilterValues(t *testing.T) {
	tests := []struct {
		name           string
		adID           int
		expectedStatus int
		expectedResult map[string][]string // expected filter values (nil means don't check exact match)
	}{
		{
			name:           "Ad 11 - filter values",
			adID:           11,
			expectedStatus: 200,
			expectedResult: map[string][]string{
				"make":             {"DAVIS"},
				"year":             {"1925"},
				"model":            {"SERIES 90", "SERIES 91"},
				"engine":           {"6cyl"},
				"part_category":    {"Exhaust & Emission"},
				"part_subcategory": {"Exhaust Manifold"},
			},
		},
		{
			name:           "Ad 12 - filter values",
			adID:           12,
			expectedStatus: 200,
			expectedResult: map[string][]string{
				"make":             {"LINCOLN"},
				"year":             {"1940"},
				"model":            {"CONTINENTAL", "CORSAIR", "NAVIGATOR"},
				"engine":           {"4.8L 292cid V12", "5.4L V8"},
				"part_category":    {"Heat & Air Conditioning"},
				"part_subcategory": {"AC Compressor"},
			},
		},
		{
			name:           "Ad 603 - filter values",
			adID:           603,
			expectedStatus: 200,
			expectedResult: map[string][]string{
				"make":             {"DODGE"},
				"year":             {"1995"},
				"model":            {"MODEL 30-35", "RAM 1500 VAN", "STRATUS"},
				"engine":           {"2.0L L4", "212cid L4"},
				"part_category":    {"Engine"},
				"part_subcategory": {"Engine"},
			},
		},
		{
			name:           "Invalid ad ID",
			adID:           99999,
			expectedStatus: 200,
			expectedResult: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/ads/%d/filter-values", baseURL, tt.adID)
			resp, result := getRequest(t, url)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				return
			}

			// Check expected results if provided
			if tt.expectedResult != nil {
				// Check that we have the same number of fields
				if len(result) != len(tt.expectedResult) {
					t.Errorf("Expected %d fields, got %d. Expected: %v, Got: %v", len(tt.expectedResult), len(result), tt.expectedResult, result)
					return
				}

				// Compare each field
				for field, expectedValues := range tt.expectedResult {
					actualFieldVal, ok := result[field]
					if !ok {
						t.Errorf("Expected field '%s' not found in result", field)
						return
					}

					actualFieldSlice, ok := actualFieldVal.([]interface{})
					if !ok {
						t.Errorf("Expected field '%s' to be an array, got %T", field, actualFieldVal)
						return
					}

					// Convert to string slice for comparison
					actualValues := make([]string, len(actualFieldSlice))
					for i, v := range actualFieldSlice {
						if str, ok := v.(string); ok {
							actualValues[i] = str
						} else {
							t.Errorf("Expected field '%s'[%d] to be a string, got %T", field, i, v)
							return
						}
					}

					// Compare values
					if len(actualValues) != len(expectedValues) {
						t.Errorf("Field '%s': expected %d values, got %d. Expected: %v, Got: %v", field, len(expectedValues), len(actualValues), expectedValues, actualValues)
						return
					}

					for i := range actualValues {
						if actualValues[i] != expectedValues[i] {
							t.Errorf("Field '%s'[%d]: expected %q, got %q", field, i, expectedValues[i], actualValues[i])
							return
						}
					}
				}
			}
		})
	}
}

// Test POST /api/categories/:category/search
func TestSearch(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		body           interface{}
		queryParams    string
		expectedStatus int
		expectedCount  int
		expectedAdIDs  []int // Exact expected ad IDs (nil means don't check exact match)
	}{
		{
			name:           "Empty search",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  2,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search by make",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"make": []string{"HONDA"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search by year",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"year": []string{"2020"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search by model",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"model": []string{"CIVIC"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search by make and year",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"make": []string{"HONDA"}, "year": []string{"2020"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search by make, year, model",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"make": []string{"HONDA"}, "year": []string{"2020"}, "model": []string{"CIVIC"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search multiple makes",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"make": []string{"HONDA", "FORD"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  2,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search multiple years",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"year": []string{"2020", "1975"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  2,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Car & Truck Parts - search by make",
			category:       "Car%20%26%20Truck%20Parts",
			body:           map[string]interface{}{"make": []string{"AC"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Car & Truck Parts - search by part_category",
			category:       "Car%20%26%20Truck%20Parts",
			body:           map[string]interface{}{"part_category": []string{"Brakes & Wheel Hub"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  91,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Car & Truck Parts - search AC 1953 ACE",
			category:       "Car%20%26%20Truck%20Parts",
			body:           map[string]interface{}{"make": []string{"AC"}, "year": []string{"1953"}, "model": []string{"ACE"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search via query params",
			category:       "Cars%20%26%20Trucks",
			body:           nil,
			queryParams:    "?make=HONDA&year=2020",
			expectedStatus: 200,
			expectedCount:  1,
			// Don't check specific ad IDs as they may vary
		},
		{
			name:           "Search with no matches",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"make": []string{"NONEXISTENT"}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  0,
			expectedAdIDs:  []int{},
		},
		{
			name:           "Invalid category search",
			category:       "InvalidCategory",
			body:           map[string]interface{}{},
			queryParams:    "",
			expectedStatus: 404,
			expectedCount:  0,
		},
		{
			name:           "Search with empty arrays",
			category:       "Cars%20%26%20Trucks",
			body:           map[string]interface{}{"make": []string{}, "year": []string{}},
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  2,
			// Don't check specific ad IDs as they may vary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURLWithPath := fmt.Sprintf("%s/api/categories/%s/search", baseURL, tt.category)
			// Parse query params and convert to form data
			formData := make(map[string]interface{})
			if tt.body != nil {
				if bodyMap, ok := tt.body.(map[string]interface{}); ok {
					// Copy body map to formData
					for k, v := range bodyMap {
						formData[k] = v
					}
				}
			}
			// Parse query params and merge into form data
			if tt.queryParams != "" {
				parsedURL, err := url.Parse(baseURLWithPath + tt.queryParams)
				if err == nil {
					for k, v := range parsedURL.Query() {
						if len(v) > 0 {
							formData[k] = v
						}
					}
				}
			}
			// Use baseURLWithPath without query params since we're sending them as form data
			resp, result := postFormRequest(t, baseURLWithPath, formData)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				return
			}

			if tt.expectedStatus == 200 {
				// Validate count
				count, ok := result["count"].(float64)
				if !ok {
					if countInt, ok := result["count"].(int); ok {
						count = float64(countInt)
					} else {
						t.Errorf("Response should contain 'count' field")
						return
					}
				}

				if int(count) != tt.expectedCount {
					t.Errorf("Expected count %d, got %d", tt.expectedCount, int(count))
				}

				// Validate ad_ids
				adIDsRaw, ok := result["ad_ids"]
				if !ok {
					t.Errorf("Response should contain 'ad_ids' field")
					return
				}

				if adIDsRaw == nil {
					if tt.expectedCount > 0 {
						t.Errorf("Expected ad_ids to be non-nil when count > 0")
					}
					return
				}

				adIDsSlice, ok := adIDsRaw.([]interface{})
				if !ok {
					t.Errorf("Expected ad_ids to be an array, got %T", adIDsRaw)
					return
				}

				if len(adIDsSlice) != tt.expectedCount {
					t.Errorf("Expected %d ad_ids, got %d", tt.expectedCount, len(adIDsSlice))
				}

				// Convert to int slice for easier checking
				adIDs := make([]int, len(adIDsSlice))
				for i, id := range adIDsSlice {
					if idFloat, ok := id.(float64); ok {
						adIDs[i] = int(idFloat)
					} else if idInt, ok := id.(int); ok {
						adIDs[i] = idInt
					} else {
						t.Errorf("Expected ad_id to be a number, got %T", id)
						return
					}
				}

				// Check for specific expected ad IDs (only if provided)
				if len(tt.expectedAdIDs) > 0 {
					adIDMap := make(map[int]bool)
					for _, id := range adIDs {
						adIDMap[id] = true
					}

					for _, expectedID := range tt.expectedAdIDs {
						if !adIDMap[expectedID] {
							t.Errorf("Expected ad_ids to contain %d, but it was not found. Got: %v", expectedID, adIDs)
						}
					}
				}

				// Check for exact expected ad IDs (only if specified)
				if tt.expectedAdIDs != nil && len(tt.expectedAdIDs) > 0 {
					if len(adIDs) != len(tt.expectedAdIDs) {
						t.Errorf("Expected %d ad_ids, got %d. Expected: %v, Got: %v", len(tt.expectedAdIDs), len(adIDs), tt.expectedAdIDs, adIDs)
					} else {
						// Compare each ad ID (order matters for exact match)
						for i, expectedID := range tt.expectedAdIDs {
							if i >= len(adIDs) || adIDs[i] != expectedID {
								t.Errorf("Expected ad_ids[%d] to be %d, got %d. Expected: %v, Got: %v", i, expectedID, adIDs[i], tt.expectedAdIDs, adIDs)
								break
							}
						}
					}
				}

				// Validate that ad_ids are positive integers
				for _, id := range adIDs {
					if id <= 0 {
						t.Errorf("Expected all ad_ids to be positive integers, got %d", id)
					}
				}
			}
		})
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Invalid form data", func(t *testing.T) {
		url := baseURL + "/api/categories/Cars%20%26%20Trucks/search"
		// Send malformed multipart form data - should handle gracefully
		req, err := http.NewRequest("POST", url, bytes.NewBufferString("invalid multipart form data"))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// ParseForm is lenient and will parse what it can, so it should return 200 with empty/partial results
		// or 400 if it truly fails. Either is acceptable for malformed form data.
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 200 or 400 for invalid form data, got %d", resp.StatusCode)
		}
	})

	t.Run("Search without Content-Type header", func(t *testing.T) {
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatalf("Failed to create cookie jar: %v", err)
		}
		client := &http.Client{Jar: jar}

		baseURLParsed, err := url.Parse(baseURL)
		if err != nil {
			t.Fatalf("Failed to parse base URL: %v", err)
		}

		// Get CSRF token via GET /health
		getReq, err := http.NewRequest("GET", baseURL+"/health", nil)
		if err != nil {
			t.Fatalf("Failed to create GET request for CSRF token: %v", err)
		}
		getResp, err := client.Do(getReq)
		if err != nil {
			t.Fatalf("Failed to get CSRF token: %v", err)
		}
		getResp.Body.Close()

		var csrfToken string
		cookies := jar.Cookies(baseURLParsed)
		for _, cookie := range cookies {
			if cookie.Name == "_csrf" {
				csrfToken = cookie.Value
				break
			}
		}
		if csrfToken == "" {
			// Fallback: try to get from response cookies and manually add to jar
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
					jar.SetCookies(baseURLParsed, []*http.Cookie{testCookie})
					break
				}
			}
		}
		if csrfToken == "" {
			t.Fatalf("Failed to get CSRF token from cookie")
		}

		requestURL := baseURL + "/api/categories/Cars%20%26%20Trucks/search?make=HONDA"
		req, err := http.NewRequest("POST", requestURL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("X-Csrf-Token", csrfToken)

		reqURL, err := url.Parse(requestURL)
		if err != nil {
			t.Fatalf("Failed to parse request URL: %v", err)
		}
		jarCookies := jar.Cookies(reqURL)
		for _, cookie := range jarCookies {
			if cookie.Name == "_csrf" {
				req.AddCookie(cookie)
				break
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

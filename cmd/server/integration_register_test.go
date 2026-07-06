package main

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/config"
)

func TestRegisterTestPhoneBypass(t *testing.T) {
	if config.GrokAPIKey == "" {
		t.Skip("GROK_API_KEY required for username screening during registration")
	}
	reflect.ValueOf(&config.AllowTestRegistration).Elem().SetBool(true)

	client := getTestClient()
	baseURLParsed, _ := url.Parse(baseURL)
	client.Jar.SetCookies(baseURLParsed, []*http.Cookie{{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}})

	suffix := time.Now().UnixNano()%9000 + 1000
	username := fmt.Sprintf("Treg%d", suffix)
	phone := fmt.Sprintf("+1555010%04d", suffix)
	password := "Ag!testpass1234"

	resp, _ := postFormRequest(t, baseURL+"/api/register/step1", map[string]interface{}{
		"username": username,
		"phone":    phone,
		"offers":   "true",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step1 status %d", resp.StatusCode)
	}

	resp, _ = postFormRequest(t, baseURL+"/api/register/step3", map[string]interface{}{
		"username":  username,
		"phone":     phone,
		"password":  password,
		"password2": password,
		"terms":     "accepted",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("step3 status %d", resp.StatusCode)
	}
	if got := resp.Header.Get("HX-Redirect"); got != "/auth/welcome" {
		t.Fatalf("expected welcome redirect, got %q", got)
	}

	resp, _ = postFormRequest(t, baseURL+"/api/login", map[string]interface{}{
		"username": username,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login after register status %d", resp.StatusCode)
	}
	hasJWT := false
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "auth_token" && cookie.Value != "" {
			hasJWT = true
			break
		}
	}
	if !hasJWT {
		t.Fatal("expected auth_token cookie after login")
	}
}

package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rocky-ads/site/internal/config"
)

func renderPageHTML(t *testing.T, title, path string) string {
	t.Helper()
	var buf bytes.Buffer
	n := Page(0, false, "", title, path, "csrf", false, nil)
	if err := n.Render(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestPageHomeSEO(t *testing.T) {
	prev := config.PublicSiteURL
	config.PublicSiteURL = "https://rockyads.com"
	t.Cleanup(func() { config.PublicSiteURL = prev })

	html := renderPageHTML(t, "Classified Ads", "/")
	wants := []string{
		"<title>Rocky Ads - Classified Ads</title>",
		`rel="canonical" href="https://rockyads.com/"`,
		`name="description" content="Rocky Ads - Classified Ads.`,
		`property="og:title" content="Rocky Ads - Classified Ads"`,
		`"@type":"WebSite"`,
		`"url":"https://rockyads.com/"`,
		`name="robots" content="index, follow"`,
	}
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Errorf("homepage missing %q", want)
		}
	}
}

func TestPageLoginSEO(t *testing.T) {
	prev := config.PublicSiteURL
	config.PublicSiteURL = "https://rockyads.com"
	t.Cleanup(func() { config.PublicSiteURL = prev })

	html := renderPageHTML(t, "Login", "/login")
	wants := []string{
		"<title>Rocky Ads - Login</title>",
		`rel="canonical" href="https://rockyads.com/login"`,
		`name="description" content="Log in to Rocky Ads`,
		`name="robots" content="index, follow"`,
	}
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Errorf("login missing %q", want)
		}
	}
	if strings.Contains(html, "application/ld+json") {
		t.Error("WebSite JSON-LD should only be on the homepage")
	}
}

func TestRobotsForPath(t *testing.T) {
	if got := robotsForPath("/auth/user/settings"); got != "noindex, nofollow" {
		t.Fatalf("auth: %q", got)
	}
	if got := robotsForPath("/about"); got != "index, follow" {
		t.Fatalf("about: %q", got)
	}
}

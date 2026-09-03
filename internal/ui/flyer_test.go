package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rocky-ads/site/internal/config"
)

func TestAdShareModalFlyer(t *testing.T) {
	path := "https://rockyads.com/ad/7"
	var buf bytes.Buffer
	if err := AdShareModal(path, "/ad/7/flyer").Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, want := range []string{
		path,
		"Ad Link",
		"Copy",
		"Flyer",
		"one-page flyer",
		`href="/ad/7/flyer"`,
		"Print flyer",
		"max-w-md",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("share modal missing %q", want)
		}
	}
	if strings.Contains(html, "flyer-thumb") {
		t.Fatal("share modal should not include flyer thumbnail")
	}

	buf.Reset()
	if err := AdShareModal(path, "").Render(&buf); err != nil {
		t.Fatal(err)
	}
	html = buf.String()
	if strings.Contains(html, "/flyer") {
		t.Fatal("inactive share modal should not include flyer")
	}
}

func TestAdFlyerPage(t *testing.T) {
	prev := config.PublicSiteURL
	config.PublicSiteURL = "https://rockyads.com"
	t.Cleanup(func() { config.PublicSiteURL = prev })

	d := AdFlyerData{
		ID:           7,
		ImageCount:   2,
		PriceDisplay: "$50",
		HasPrice:     true,
		Title:        "Blue Bicycle",
		Location:     "Denver, CO",
		Description:  "Gently used.",
		FacetLabel:   "2K mi",
		FacetDetails: []string{"Year: 2018"},
		Tags:         []string{"road"},
		AdURL:        "https://rockyads.com/ad/7",
	}
	var buf bytes.Buffer
	if err := AdFlyerPage(d).Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, want := range []string{
		"Blue Bicycle",
		"$50",
		"Denver, CO",
		"Gently used.",
		"2K mi",
		"Year: 2018",
		"road",
		"https://rockyads.com/ad/7",
		"data:image/png;base64,",
		"Print",
		`href="/ad/7"`,
		"noindex, nofollow",
		"window.print()",
		"flyer-screen-only",
		"flyer-sheet",
		"size: letter",
		"max-height: 10in",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("flyer missing %q", want)
		}
	}
	if strings.Contains(html, `class="text-xl font-bold"`) {
		t.Fatal("did not expect Rocky Ads heading on flyer")
	}
	descAt := strings.Index(html, "Gently used.")
	qrAt := strings.Index(html, "QR code for this ad")
	if descAt < 0 || qrAt < 0 || descAt > qrAt {
		t.Fatal("description should sit in the text column before the QR")
	}
}

func TestAdFlyerPageNoImages(t *testing.T) {
	d := AdFlyerData{
		ID:    3,
		Title: "No Photos",
		AdURL: "https://rockyads.com/ad/3",
	}
	var buf bytes.Buffer
	if err := AdFlyerPage(d).Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	if strings.Contains(html, "Ad image") {
		t.Fatal("did not expect image markup with ImageCount 0")
	}
}

func TestAdFlyerPageCapsImages(t *testing.T) {
	d := AdFlyerData{
		ID:         7,
		ImageCount: 9,
		Title:      "Many Photos",
		AdURL:      "https://rockyads.com/ad/7",
	}
	var buf bytes.Buffer
	if err := AdFlyerPage(d).Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	if !strings.Contains(html, "ad: 7, image: 4") {
		t.Fatal("expected fourth flyer image")
	}
	if strings.Contains(html, "ad: 7, image: 5") {
		t.Fatal("flyer should cap at 4 images")
	}
}

func TestAdFlyerPageThreeImagesTwoByTwo(t *testing.T) {
	d := AdFlyerData{
		ID:         7,
		ImageCount: 3,
		Title:      "Three Photos",
		AdURL:      "https://rockyads.com/ad/7",
	}
	var buf bytes.Buffer
	if err := AdFlyerPage(d).Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	if !strings.Contains(html, "repeat(2, minmax(0, 1fr))") {
		t.Fatal("three images should use a 2-column grid")
	}
	if strings.Contains(html, "repeat(3,") {
		t.Fatal("three images should not use a 3-column grid")
	}
}

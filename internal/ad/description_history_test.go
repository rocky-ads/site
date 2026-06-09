package ad

import (
	"strings"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/facet"
)

func TestSplitDescription(t *testing.T) {
	original := "Triumph tr7 $3500 obo."
	history := historyMarker + "1/7/2026 12:34 am  Description Addition\n\nHas the 2.0L engine."
	desc := original + historyEndMarker + history

	gotOrig, gotHist := SplitDescription(desc)
	if gotOrig != original {
		t.Errorf("original = %q, want %q", gotOrig, original)
	}
	if gotHist != history {
		t.Errorf("history = %q, want %q", gotHist, history)
	}

	gotOrig, gotHist = SplitDescription("no history here")
	if gotOrig != "no history here" || gotHist != "" {
		t.Errorf("no marker split failed: orig=%q hist=%q", gotOrig, gotHist)
	}
}

func TestAppendHistoryEntryPrependsNewest(t *testing.T) {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	first := time.Date(2026, 1, 7, 0, 34, 0, 0, loc)
	second := time.Date(2026, 1, 7, 18, 47, 0, 0, loc)

	desc := AppendHistoryEntry(
		"Original text.",
		"Description Addition",
		"First edit.",
		first,
		loc,
	)
	desc = AppendHistoryEntry(
		desc,
		"Price change",
		"Price dropped from $3,400 to $3,000",
		second,
		loc,
	)

	parts := ParseDescriptionForDisplay(desc)
	if parts.Original != "Original text." {
		t.Errorf("original = %q", parts.Original)
	}
	if len(parts.History) != 2 {
		t.Fatalf("got %d history entries, want 2", len(parts.History))
	}
	if !strings.Contains(parts.History[0].Header, "Price change") {
		t.Errorf("newest first: %q", parts.History[0].Header)
	}
	if !strings.Contains(parts.History[1].Header, "Description Addition") {
		t.Errorf("oldest second: %q", parts.History[1].Header)
	}
	if !strings.Contains(desc, historyEndMarker) {
		t.Fatal("expected history end marker")
	}
}

func TestAppendHistoryEntry(t *testing.T) {
	loc, _ := time.LoadLocation("America/Los_Angeles")
	at := time.Date(2026, 1, 7, 0, 34, 0, 0, loc)

	desc := AppendHistoryEntry(
		"Original text.",
		"Description Addition",
		"Has the 2.0L engine.",
		at,
		loc,
	)
	if !strings.Contains(desc, historyMarker) {
		t.Fatal("expected history marker in stored description")
	}
	orig, _ := SplitDescription(desc)
	if strings.Contains(orig, historyMarker) {
		t.Fatal("marker must not appear in original portion")
	}
	display := DisplayDescription(desc)
	if strings.Contains(display, historyMarker) {
		t.Fatal("marker visible after DisplayDescription")
	}
	if !strings.Contains(display, "1/7/2026 12:34 am  Description Addition") {
		t.Errorf("display missing header: %q", display)
	}
	if !strings.Contains(display, "Has the 2.0L engine.") {
		t.Errorf("display missing body: %q", display)
	}
}

func TestBuildFieldChangeEntries(t *testing.T) {
	oldAmt, newAmt := 3400, 3000
	code := "USD"
	oldAd := Ad{
		Title: "Old title",
		Facets: map[string]facet.Value{
			"price": {Num: &oldAmt, Text: &code},
		},
	}
	newFacets := map[string]facet.Value{
		"price": {Num: &newAmt, Text: &code},
	}
	cat := Category{FacetKeys: []string{"price"}}

	entries := BuildFieldChangeEntries(
		oldAd, "New title", "", newFacets, cat,
	)
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].label != "Title change" {
		t.Errorf("first label = %q", entries[0].label)
	}
	if entries[1].label != "Price change" {
		t.Errorf("second label = %q", entries[1].label)
	}
	if !strings.Contains(entries[1].body, "Price dropped") {
		t.Errorf("price body = %q", entries[1].body)
	}
}

func TestFormatLocationHistoryChangeNoOp(t *testing.T) {
	old := Ad{
		RawLocation: "97333",
		City:        "Corvallis",
		AdminArea:   "OR",
		Country:     "US",
	}
	if body := formatLocationHistoryChange(old, "97333", Category{}); body != "" {
		t.Fatalf("want no change, got %q", body)
	}
}

func TestSanitizeStripsHistoryMarker(t *testing.T) {
	if got := SanitizeAdText("before\u001eafter"); got != "beforeafter" {
		t.Errorf("SanitizeAdText = %q", got)
	}
	if got := SanitizeAdText("before\u001fafter"); got != "beforeafter" {
		t.Errorf("SanitizeAdText end marker = %q", got)
	}
}

func TestFormatLocationHistoryChangeAddressCategory(t *testing.T) {
	addr := "123 Main St"
	old := Ad{
		Facets: map[string]facet.Value{
			"address": {Text: &addr},
		},
	}
	garage := Category{FacetKeys: []string{"address"}}

	if body := formatLocationHistoryChange(old, "456 Oak Ave", garage); body != "Address changed" {
		t.Fatalf("change = %q, want Address changed", body)
	}
	if body := formatLocationHistoryChange(old, "", garage); body != "Address removed" {
		t.Fatalf("remove = %q, want Address removed", body)
	}
	if body := formatLocationHistoryChange(Ad{}, "123 Main St", garage); body != "Address changed" {
		t.Fatalf("set = %q, want Address changed", body)
	}
	if body := formatLocationHistoryChange(old, addr, garage); body != "" {
		t.Fatalf("no-op = %q, want empty", body)
	}
}

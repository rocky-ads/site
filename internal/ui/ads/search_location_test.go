package ads

import "testing"

func TestLocationSummaryText(t *testing.T) {
	t.Run("no location", func(t *testing.T) {
		got := LocationSummaryText(SearchFilters{})
		if got != "No location set" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("with resolved display", func(t *testing.T) {
		got := LocationSummaryText(SearchFilters{
			Location:        "97345",
			LocationDisplay: "\U0001F1FA\U0001F1F8 Corvallis, OR",
			Within:          50,
			WithinUnit:      "mi",
		})
		want := "\U0001F1FA\U0001F1F8 Corvallis, OR - Within 50 miles"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
	t.Run("falls back to raw location", func(t *testing.T) {
		got := LocationSummaryText(SearchFilters{
			Location:   "Corvallis, Oregon",
			Within:     50,
			WithinUnit: "mi",
		})
		want := "Corvallis, Oregon - Within 50 miles"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

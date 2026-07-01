package seed

import (
	"strings"
	"testing"

	"github.com/rocky-ads/site/internal/ad"
)

func TestAssembleDescription(t *testing.T) {
	history := []HistoryEntryJSON{
		{
			Label: "Description Addition",
			Body:  "Available for local pickup only.",
			At:    "2024-09-12T15:30:00-04:00",
		},
	}
	got, err := AssembleDescription("Original body.", history)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "\u001f") {
		t.Fatal("expected history end marker")
	}
	parts := ad.ParseDescriptionForDisplay(got)
	if parts.Original != "Original body." {
		t.Errorf("original = %q", parts.Original)
	}
	if len(parts.History) != 1 {
		t.Fatalf("got %d history entries", len(parts.History))
	}
}

func TestSeedTagsJSON(t *testing.T) {
	aj := adJSON{
		Suggestions: []SuggestionJSON{
			{Label: "OEM", Value: "yes"},
			{Label: "side", Value: "left"},
		},
	}
	raw := ad.TagsJSON(seedSuggestions(aj))
	if raw == "[]" {
		t.Fatal("expected tags json")
	}
	if !strings.Contains(raw, "OEM") {
		t.Fatalf("unexpected json: %s", raw)
	}
}

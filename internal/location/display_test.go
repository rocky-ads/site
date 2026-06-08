package location

import (
	"strings"
	"testing"
)

func TestDisplayText(t *testing.T) {
	got := DisplayText("Klamath Falls", "OR", "US")
	if got == "" {
		t.Fatal("expected non-empty display")
	}
	if !strings.Contains(got, "Klamath Falls") ||
		!strings.Contains(got, "OR") {
		t.Fatalf("unexpected display: %q", got)
	}
}

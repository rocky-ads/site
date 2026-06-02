package field

import "testing"

func TestParseInputType(t *testing.T) {
	tests := []struct {
		raw      string
		wantType string
		wantMax  string
		wantMin  string
		wantPat  string
	}{
		{"select", InputSelect, "", "", ""},
		{"select_multi", InputSelectMulti, "", "", ""},
		{"text?max=35&pattern=ascii", InputText, "35", "", "ascii"},
		{"type=text&max=32&pattern=ascii", InputText, "32", "", "ascii"},
		{"number?min=0&pattern=nonneg-int", InputNumber, "", "0", "nonneg-int"},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			spec := ParseInputType(tt.raw)
			if spec.Type != tt.wantType {
				t.Fatalf("Type = %q, want %q", spec.Type, tt.wantType)
			}
			if got := spec.Param("max"); got != tt.wantMax {
				t.Fatalf("max = %q, want %q", got, tt.wantMax)
			}
			if got := spec.Param("min"); got != tt.wantMin {
				t.Fatalf("min = %q, want %q", got, tt.wantMin)
			}
			if got := spec.Param("pattern"); got != tt.wantPat {
				t.Fatalf("pattern = %q, want %q", got, tt.wantPat)
			}
		})
	}
}

func TestInputSpec_patternShorthands(t *testing.T) {
	tests := []struct {
		pattern string
		wantRE  string
		wantMsg string
	}{
		{"ascii", `[\x20-\x7E]+`, "Please enter printable ASCII characters only"},
		{"ascii-multiline", `[\x20-\x7E\n\r]+`, "Please enter printable ASCII characters only (line breaks allowed)"},
		{"nonneg-int", `^(0|[1-9][0-9]*)$`, "Please enter a whole number (0 or greater)"},
	}

	for _, tt := range tests {
		spec := ParseInputType("text?pattern=" + tt.pattern)
		if got := spec.HTMLPattern(); got != tt.wantRE {
			t.Fatalf("HTMLPattern(%s) = %q, want %q", tt.pattern, got, tt.wantRE)
		}
		if got := spec.PatternMessage(); got != tt.wantMsg {
			t.Fatalf("PatternMessage(%s) = %q, want %q", tt.pattern, got, tt.wantMsg)
		}
	}
}

func TestAnchoredPattern(t *testing.T) {
	if got := AnchoredPattern(`[\x20-\x7E\n\r]+`); got != `^(?:[\x20-\x7E\n\r]+)$` {
		t.Fatalf("AnchoredPattern = %q", got)
	}
	if got := AnchoredPattern(`^(0|[1-9][0-9]*)$`); got != `^(0|[1-9][0-9]*)$` {
		t.Fatalf("already anchored = %q", got)
	}
}

func TestIsMultiInput_parsed(t *testing.T) {
	if !IsMultiInput("select_multi") {
		t.Fatal("expected select_multi")
	}
	if IsMultiInput("text?max=10") {
		t.Fatal("expected text not multi")
	}
}

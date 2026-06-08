package ad

import (
	"strings"
	"testing"
	"time"
)

type stubCompressor struct {
	fn func(systemPrompt, text string, maxRunes int) (string, error)
}

func (s stubCompressor) Compress(
	systemPrompt, text string, maxRunes int,
) (string, error) {
	return s.fn(systemPrompt, text, maxRunes)
}

func TestEnsureDescriptionFitsNoOp(t *testing.T) {
	desc := "short description"
	got, err := ensureDescriptionFits(
		desc, time.Now(), time.UTC, stubCompressor{
			fn: func(_, _ string, _ int) (string, error) {
				t.Fatal("compressor should not be called")
				return "", nil
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != desc {
		t.Errorf("got %q", got)
	}
}

func TestEnsureDescriptionFitsCompressesOriginal(t *testing.T) {
	original := strings.Repeat("word ", 300)
	desc := original
	calls := 0
	got, err := ensureDescriptionFits(
		desc, time.Now(), time.UTC, stubCompressor{
			fn: func(prompt, text string, maxRunes int) (string, error) {
				calls++
				if calls == 1 {
					if text != original {
						t.Errorf("first compress text mismatch")
					}
					return "summary", nil
				}
				return "hist summary", nil
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "summary") {
		t.Errorf("got %q", got)
	}
	if !strings.Contains(got, "Description compressed") {
		t.Errorf("missing compression note: %q", got)
	}
}

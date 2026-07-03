package vector_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rocky-ads/site/internal/vector"
)

func TestIsEmbeddingUnavailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"no embedded ads", vector.ErrNoEmbeddedAds, true},
		{"rate limit", fmt.Errorf("Gemini embedding API: Error 429"), true},
		{"resource exhausted",
			fmt.Errorf("RESOURCE_EXHAUSTED"), true},
		{"other", errors.New("database down"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vector.IsEmbeddingUnavailable(tt.err); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

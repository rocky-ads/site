package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmbeddingsTabFilterTargetsListings(t *testing.T) {
	var buf bytes.Buffer
	if err := EmbeddingsTab(EmbeddingAdminData{}).Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	wants := []string{
		`id="embedding-user-activity-rows"`,
		`id="embedding-site-activity-rows"`,
		`hx-get="/admin/embeddings/user-activities"`,
		`hx-get="/admin/embeddings/site-activities"`,
		`hx-target="#embedding-user-activity-rows"`,
		`hx-target="#embedding-site-activity-rows"`,
	}
	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q", want)
		}
	}
}

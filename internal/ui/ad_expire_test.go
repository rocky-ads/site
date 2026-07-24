package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	g "maragu.dev/gomponents"
)

func TestFormatExpiresIn(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{"soon_minutes", 30 * time.Minute, "Expires soon"},
		{"soon_hours", 5 * time.Hour, "Expires soon"},
		{"one_day", 25 * time.Hour, "Expires in 1 day"},
		{"days", 5*24*time.Hour + time.Hour, "Expires in 5 days"},
		{"one_month", 30 * 24 * time.Hour, "Expires in 1 month"},
		{"one_month_days", 42 * 24 * time.Hour, "Expires in 1 month 12 days"},
		{"months", 60 * 24 * time.Hour, "Expires in 2 months"},
		{"months_days", 72 * 24 * time.Hour, "Expires in 2 months 12 days"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add a second so time.Until does not slip under a day boundary.
			got := formatExpiresIn(time.Now().Add(tt.in + time.Second))
			if got != tt.want {
				t.Fatalf("formatExpiresIn(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestAdExpireToolbarVisibility(t *testing.T) {
	created := time.Now().UTC()
	detail := AdDetail{
		ID:        1,
		OwnerID:   42,
		Title:     "Test Ad",
		Active:    true,
		CreatedAt: created,
	}

	t.Run("owner_active", func(t *testing.T) {
		html := renderAdNodes(t, Ad(detail, 42, "csrf"))
		if !strings.Contains(html, "Expires in") {
			t.Fatal("expected expire countdown for owner")
		}
		if !strings.Contains(html, "/images/post_add.svg") {
			t.Fatal("expected new-ad icon for owner")
		}
		if !strings.Contains(html, "/images/copy.svg") {
			t.Fatal("expected copy-ad icon for owner")
		}
		if !strings.Contains(html, `/auth/ad/1/copy`) {
			t.Fatal("expected copy-ad link for owner")
		}
	})

	t.Run("non_owner", func(t *testing.T) {
		html := renderAdNodes(t, Ad(detail, 99, "csrf"))
		if strings.Contains(html, "Expires in") {
			t.Fatal("did not expect expire countdown for non-owner")
		}
		if strings.Contains(html, "/images/post_add.svg") {
			t.Fatal("did not expect new-ad icon for non-owner")
		}
		if strings.Contains(html, "/images/copy.svg") {
			t.Fatal("did not expect copy-ad icon for non-owner")
		}
	})

	t.Run("owner_inactive", func(t *testing.T) {
		paused := detail
		paused.Active = false
		paused.Inactive = true
		html := renderAdNodes(t, Ad(paused, 42, "csrf"))
		if strings.Contains(html, "Expires in") {
			t.Fatal("did not expect expire countdown for paused ad")
		}
		if strings.Contains(html, "/images/post_add.svg") {
			t.Fatal("did not expect new-ad icon for paused ad")
		}
		if strings.Contains(html, "/images/copy.svg") {
			t.Fatal("did not expect copy-ad icon for paused ad")
		}
	})
}

func renderAdNodes(t *testing.T, nodes []g.Node) string {
	t.Helper()
	var buf bytes.Buffer
	for _, n := range nodes {
		if err := n.Render(&buf); err != nil {
			t.Fatalf("render: %v", err)
		}
	}
	return buf.String()
}

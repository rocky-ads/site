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
		{"soon", 30 * time.Second, "Expires soon"},
		{"one_minute", time.Minute + time.Second, "Expires in 1 minute"},
		{"minutes", 12*time.Minute + time.Second, "Expires in 12 minutes"},
		{"one_hour", time.Hour + time.Minute, "Expires in 1 hour"},
		{"hours", 5*time.Hour + time.Minute, "Expires in 5 hours"},
		{"one_day", 25 * time.Hour, "Expires in 1 day"},
		{"days", 5*24*time.Hour + time.Hour, "Expires in 5 days"},
		{"one_month", 35 * 24 * time.Hour, "Expires in 1 month"},
		{"months", 70 * 24 * time.Hour, "Expires in 2 months"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatExpiresIn(time.Now().Add(tt.in))
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

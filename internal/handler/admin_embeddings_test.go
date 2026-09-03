package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestFormatActivityTime(t *testing.T) {
	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		raw  string
		loc  *time.Location
		want string
	}{
		{"", time.UTC, "—"},
		{"—", time.UTC, "—"},
		{"2026-09-02T23:39:21.644426", la, "Sep 2, 4:39 PM"},
		{"2026-09-02T23:39:21.644426", time.UTC, "Sep 2, 11:39 PM"},
		{"2026-09-02T16:39:21-07:00", la, "Sep 2, 4:39 PM"},
	}
	for _, tt := range tests {
		got := formatActivityTime(tt.raw, tt.loc)
		if got != tt.want {
			t.Errorf("formatActivityTime(%q) = %q, want %q",
				tt.raw, got, tt.want)
		}
	}
}

func TestEmbeddingUserID(t *testing.T) {
	app := fiber.New()
	var got int
	app.Get("/t", func(c *fiber.Ctx) error {
		got = embeddingUserID(c, 1)
		return nil
	})
	tests := []struct {
		url  string
		want int
	}{
		{"/t", 1},
		{"/t?user=6", 6},
		{"/t?user=nope", 1},
		{"/t?user=0", 1},
	}
	for _, tt := range tests {
		if _, err := app.Test(httptest.NewRequest("GET", tt.url, nil)); err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.url, got, tt.want)
		}
	}
}

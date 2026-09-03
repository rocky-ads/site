package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func visitorIPFromReq(t *testing.T, headers map[string]string) string {
	t.Helper()
	app := fiber.New()
	var got string
	app.Get("/", func(c *fiber.Ctx) error {
		got = VisitorIP(c)
		return c.SendStatus(fiber.StatusOK)
	})
	req := httptest.NewRequest(fiber.MethodGet, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	return got
}

func TestVisitorIPPrefersCFConnectingIP(t *testing.T) {
	got := visitorIPFromReq(t, map[string]string{
		"CF-Connecting-IP": "203.0.113.10",
		"True-Client-IP":   "198.51.100.20",
		"X-Forwarded-For":  "198.51.100.1, 10.1.1.1",
	})
	if got != "203.0.113.10" {
		t.Fatalf("got %q", got)
	}
}

func TestVisitorIPUsesTrueClientIP(t *testing.T) {
	got := visitorIPFromReq(t, map[string]string{
		"True-Client-IP":  "198.51.100.20",
		"X-Forwarded-For": "203.0.113.10, 10.1.1.1",
	})
	if got != "198.51.100.20" {
		t.Fatalf("got %q", got)
	}
}

func TestVisitorIPUsesFirstPublicForwarded(t *testing.T) {
	got := visitorIPFromReq(t, map[string]string{
		"X-Forwarded-For": "10.192.43.130, 203.0.113.10, 10.1.1.1",
	})
	if got != "203.0.113.10" {
		t.Fatalf("got %q", got)
	}
}

func TestVisitorIPSkipsPrivateEdgeHeaders(t *testing.T) {
	got := visitorIPFromReq(t, map[string]string{
		"CF-Connecting-IP": "10.228.17.159",
		"X-Forwarded-For":  "203.0.113.10",
	})
	if got != "203.0.113.10" {
		t.Fatalf("got %q", got)
	}
}

func TestVisitorIPFallsBackToPeer(t *testing.T) {
	got := visitorIPFromReq(t, nil)
	if got == "" {
		t.Fatal("expected peer IP")
	}
	if publicIP(got) != "" {
		t.Fatalf("expected private/loopback peer, got %q", got)
	}
}

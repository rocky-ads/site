package cookie

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSetCategoryIDNotOverwrittenByLaterGet(t *testing.T) {
	app := fiber.New()
	app.Get("/c/:id", func(c *fiber.Ctx) error {
		// /c/:id sets the deep-linked category, then other code reads
		// GetCategoryID (as switchCategoryHome does). A first-time visitor
		// has no request cookie; Get must not reset the response cookie.
		SetCategoryID(c, 4)
		got := GetCategoryID(c)
		if got != 4 {
			t.Fatalf("GetCategoryID = %d, want 4", got)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/c/4", nil))
	if err != nil {
		t.Fatal(err)
	}

	var category string
	for _, c := range resp.Cookies() {
		if c.Name == "category" {
			category = c.Value
		}
	}
	if category != "4" {
		t.Fatalf("category cookie = %q, want %q (Set-Cookie: %s)",
			category, "4", resp.Header.Get("Set-Cookie"))
	}
	if !strings.Contains(strings.ToLower(resp.Header.Get("Set-Cookie")), "samesite=lax") {
		t.Fatalf("expected SameSite=Lax for QR deep links, got %q",
			resp.Header.Get("Set-Cookie"))
	}
}

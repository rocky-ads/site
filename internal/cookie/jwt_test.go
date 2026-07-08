package cookie

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestClearJWT(t *testing.T) {
	app := fiber.New()
	var clearHeader string

	app.Post("/clear", func(c *fiber.Ctx) error {
		ClearJWT(c)
		clearHeader = string(c.Response().Header.Peek("Set-Cookie"))
		return c.SendStatus(fiber.StatusOK)
	})

	if _, err := app.Test(httptest.NewRequest(http.MethodPost, "/clear", nil)); err != nil {
		t.Fatal(err)
	}

	clearHeader = strings.ToLower(clearHeader)
	for _, want := range []string{
		"auth_token=",
		"path=/",
		"httponly",
		"samesite=strict",
		"max-age=0",
	} {
		if !strings.Contains(clearHeader, want) {
			t.Errorf("Set-Cookie = %q, want substring %q", clearHeader, want)
		}
	}
}

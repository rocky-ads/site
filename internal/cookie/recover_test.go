package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRecoverSessionCookieRoundTrip(t *testing.T) {
	app := fiber.New()
	app.Get("/set", func(c *fiber.Ctx) error {
		SetRecoverSession(c, "tok_abc", "123456")
		return c.SendStatus(fiber.StatusOK)
	})
	var token, code string
	app.Get("/get", func(c *fiber.Ctx) error {
		token = GetRecoverSession(c)
		code = GetRecoverCode(c)
		return c.SendStatus(fiber.StatusOK)
	})

	setReq := httptest.NewRequest(http.MethodGet, "/set", nil)
	setResp, err := app.Test(setReq)
	if err != nil {
		t.Fatal(err)
	}
	getReq := httptest.NewRequest(http.MethodGet, "/get", nil)
	for _, c := range setResp.Cookies() {
		getReq.AddCookie(c)
	}
	if _, err := app.Test(getReq); err != nil {
		t.Fatal(err)
	}
	if token != "tok_abc" || code != "123456" {
		t.Fatalf("got token=%q code=%q", token, code)
	}
}

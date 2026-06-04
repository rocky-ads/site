package cookie

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSearchStateRoundTrip(t *testing.T) {
	min := 100
	max := 5000
	want := SearchState{
		Q:        "Honda",
		PriceMin: &min,
		PriceMax: &max,
		Location: "Denver",
		Radius:   25,
		Expanded: true,
	}
	app := fiber.New()
	var got SearchState
	app.Get("/set", func(c *fiber.Ctx) error {
		SetSearchState(c, want)
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/get", func(c *fiber.Ctx) error {
		got = GetSearchState(c)
		return c.SendStatus(fiber.StatusOK)
	})

	setResp, err := app.Test(httptest.NewRequest("GET", "/set", nil))
	if err != nil {
		t.Fatal(err)
	}
	var cookieVal string
	for _, c := range setResp.Cookies() {
		if c.Name == searchCookieName {
			cookieVal = c.Value
			break
		}
	}
	if cookieVal == "" {
		t.Fatal("expected search cookie in response")
	}

	getReq := httptest.NewRequest("GET", "/get", nil)
	getReq.Header.Set("Cookie", searchCookieName+"="+cookieVal)
	if _, err := app.Test(getReq); err != nil {
		t.Fatal(err)
	}
	if got.Q != want.Q || got.Location != want.Location ||
		got.Radius != want.Radius || got.Expanded != want.Expanded {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if got.PriceMin == nil || *got.PriceMin != min {
		t.Fatalf("price_min: got %v", got.PriceMin)
	}
}

func TestGetSearchStateInvalidCookie(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		state := GetSearchState(c)
		if state.Q != "" || state.Expanded {
			t.Fatalf("expected zero state, got %+v", state)
		}
		return c.SendStatus(fiber.StatusOK)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Cookie", searchCookieName+"=not-valid-base64!!!")
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
}

func TestSearchStateJSON(t *testing.T) {
	min := 10
	state := SearchState{Q: "test", PriceMin: &min, Expanded: true}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	if _, err := base64.RawURLEncoding.DecodeString(encoded); err != nil {
		t.Fatal(err)
	}
}

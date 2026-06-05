package handler

import (
	"encoding/base64"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/search"
)

func TestParseSearchParamsFromStateMileage(t *testing.T) {
	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(":memory:"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO categories (name, seed_ad_file, image_file, flags) VALUES ('Car & Truck Parts', 'p.json', 'c.svg', 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO categories (name, seed_ad_file, image_file, flags) VALUES ('Cars & Trucks', 'a.json', 'c.svg', 1)`); err != nil {
		t.Fatal(err)
	}
	if err := ad.LoadCategories(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (encrypted_name, name_nonce, name_hash, password_hash, password_salt, encrypted_phone, phone_nonce, phone_hash, encrypted_email, email_nonce)
		VALUES (x'', x'', 'h', 'p', 's', x'', x'', 'ph', x'', x'')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ads (category_id, title, description, price, user_id, mileage) VALUES (2, 'Higher miles car', 'desc', 100, 1, 45000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ads (category_id, title, description, price, user_id, mileage) VALUES (2, 'Lower miles car', 'desc', 100, 1, 28000)`); err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	var params search.Params
	app.Get("/search", func(c *fiber.Ctx) error {
		state := saveSearchStateFromRequest(c, nil, true)
		params = parseSearchParamsFromState(c, state, 2)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/search?mileage_min=40000&mileage_max=50000", nil)
	req.Header.Set("Cookie", "search="+encodeSearchState(`{"expanded":true}`))

	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}

	ids, err := search.Search(params)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 ad, got %v params=%+v", ids, params)
	}
}

func encodeSearchState(json string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(json))
}

package handler

import (
	"encoding/base64"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/facet"
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
	if _, err := db.Exec(`INSERT INTO categories (name, seed_ad_file, image_file, facets) VALUES ('Car & Truck Parts', 'p.json', 'c.svg', '["price"]')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO categories (name, seed_ad_file, image_file, facets) VALUES ('Cars & Trucks', 'a.json', 'c.svg', '["mileage","price"]')`); err != nil {
		t.Fatal(err)
	}
	if err := ad.LoadCategories(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (encrypted_name, name_nonce, name_hash, password_hash, password_salt, encrypted_phone, phone_nonce, phone_hash, encrypted_email, email_nonce)
		VALUES (x'', x'', 'h', 'p', 's', x'', x'', 'ph', x'', x'')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ads (id, category_id, title, description, user_id) VALUES (1, 2, 'Higher miles car', 'desc', 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES (1, 'mileage', 45000, 'mi')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ads (id, category_id, title, description, user_id) VALUES (2, 2, 'Lower miles car', 'desc', 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES (2, 'mileage', 28000, 'mi')`); err != nil {
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
	req.Header.Set("Cookie", "category=2; search="+encodeSearchState(`{"expanded":true}`))

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

func TestParseFacetFiltersConditionCheckboxes(t *testing.T) {
	category := ad.Category{FacetKeys: []string{"condition"}}

	app := fiber.New()
	var filters map[string]facet.Filter
	app.Get("/search", func(c *fiber.Ctx) error {
		filters = parseFacetFilters(c, category)
		return c.SendStatus(fiber.StatusOK)
	})

	q := url.Values{}
	q.Add("condition", "New")
	q.Add("condition", "Used - Good")
	req := httptest.NewRequest("GET", "/search?"+q.Encode(), nil)
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
	got := filters["condition"]
	if len(got.Values) != 2 || got.Values[0] != "New" || got.Values[1] != "Used - Good" {
		t.Fatalf("condition filter = %+v", got)
	}
}

func encodeSearchState(json string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(json))
}

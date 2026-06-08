package search

import (
	"os"
	"testing"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/location"
)

func TestSearchGeoFilterSQL(t *testing.T) {
	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(":memory:"); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO locations (raw_text, city, admin_area, country, latitude, longitude)
		VALUES ('Denver', 'Denver', 'CO', 'US', 39.7392, -104.9903)`)
	if err != nil {
		t.Fatal(err)
	}
	var locID int
	_ = db.QueryRow(`SELECT id FROM locations`).Scan(&locID)
	_, err = db.Exec(`INSERT INTO users (encrypted_name, name_nonce, name_hash, password_hash, password_salt, encrypted_phone, phone_nonce, phone_hash, encrypted_email, email_nonce)
		VALUES (x'', x'', 'h', 'p', 's', x'', x'', 'ph', x'', x'')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO categories (name, seed_ad_file, image_file) VALUES ('Test', 'a.json', 'c.svg')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO ads (category_id, title, description, user_id, location_id)
		VALUES (1, 'Near Denver Bike', 'desc', 1, ?)`, locID)
	if err != nil {
		t.Fatal(err)
	}

	lat, lon, ok, err := location.ResolveLocation("Denver")
	if err != nil || !ok {
		t.Fatalf("resolve: ok=%v err=%v", ok, err)
	}

	p := Params{
		CategoryID: 1,
		Limit:      10,
		CenterLat:  lat,
		CenterLon:  lon,
		RadiusKm:   location.MilesToKm(50),
		HasGeo:     true,
	}
	ids, err := Search(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 ad, got %v", ids)
	}

	t.Run("condition checkboxes", func(t *testing.T) {
		insertAdWithCondition(t, "New bike", "New")
		insertAdWithCondition(t, "Fair bike", "Used - Fair")
		insertAdWithCondition(t, "Poor bike", "Used - Poor")
		ids, err := Search(Params{
			CategoryID: 1,
			Limit:      10,
			FacetFilters: map[string]facet.Filter{"condition": {
				Values: []string{"New", "Used - Fair"},
			}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != 2 {
			t.Fatalf("expected 2 ads, got %v", ids)
		}
	})
	t.Run("mileage range", func(t *testing.T) {
		insertAdWithMileage(t, "Low miles car", 28000)
		insertAdWithMileage(t, "Higher miles car", 45000)
		min := 40000
		max := 50000
		ids, err := Search(Params{
			CategoryID:   1,
			Limit:        10,
			Offset:       0,
			FacetFilters: map[string]facet.Filter{"mileage": {Min: &min, Max: &max}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != 1 {
			t.Fatalf("expected 1 ad, got %v", ids)
		}
	})
}

func insertAdWithCondition(t *testing.T, title, condition string) {
	t.Helper()
	res, err := db.Exec(`INSERT INTO ads (category_id, title, description, user_id)
		VALUES (1, ?, 'desc', 1)`, title)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES (?, 'condition', NULL, ?)`, id, condition); err != nil {
		t.Fatal(err)
	}
}

func insertAdWithMileage(t *testing.T, title string, mileage int) {
	t.Helper()
	res, err := db.Exec(`INSERT INTO ads (category_id, title, description, user_id)
		VALUES (1, ?, 'desc', 1)`, title)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES (?, 'mileage', ?, 'mi')`, id, mileage); err != nil {
		t.Fatal(err)
	}
}

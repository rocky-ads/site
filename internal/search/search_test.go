package search

import (
	"os"
	"testing"

	"github.com/rocky-ads/site/internal/db"
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
		VALUES ('Denver, CO, US', 'Denver', 'CO', 'US', 39.7392, -104.9903)`)
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
	_, err = db.Exec(`INSERT INTO ads (category_id, title, description, price, user_id, location_id)
		VALUES (1, 'Near Denver Bike', 'desc', 100, 1, ?)`, locID)
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
}

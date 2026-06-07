package ad

import (
	"os"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/db"
)

func TestGetAdNullLocationID(t *testing.T) {
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
		VALUES (1, 'No location ad', 'desc', 1, NULL)`)
	if err != nil {
		t.Fatal(err)
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")
	got, err := GetAd(0, 1, loc)
	if err != nil {
		t.Fatalf("GetAd(1) failed: %v", err)
	}
	if got.LocationID != nil {
		t.Fatalf("expected nil location_id, got %v", *got.LocationID)
	}
}

package location

import (
	"os"
	"testing"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
)

func TestParseLocationResponse(t *testing.T) {
	lat := 44.5646
	lon := -123.2620
	resp := `{"city":"Corvallis","admin_area":"OR","country":"US",
"latitude":44.5646,"longitude":-123.2620}`
	got, err := parseLocationResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	if got.City != "Corvallis" || got.AdminArea != "OR" ||
		got.Country != "US" {
		t.Fatalf("unexpected location: %+v", got)
	}
	if got.Latitude == nil || *got.Latitude != lat {
		t.Fatalf("latitude = %v", got.Latitude)
	}
	if got.Longitude == nil || *got.Longitude != lon {
		t.Fatalf("longitude = %v", got.Longitude)
	}
}

func TestParseLocationResponseStripsFence(t *testing.T) {
	resp := "```json\n" +
		`{"city":"Portland","admin_area":"OR","country":"US",
"latitude":45.5152,"longitude":-122.6784}` +
		"\n```"
	got, err := parseLocationResponse(resp)
	if err != nil {
		t.Fatal(err)
	}
	if got.City != "Portland" {
		t.Fatalf("city = %q", got.City)
	}
}

func TestParseLocationResponseMissingCoords(t *testing.T) {
	resp := `{"city":"Nowhere","admin_area":"OR","country":"US"}`
	_, err := parseLocationResponse(resp)
	if err == nil {
		t.Fatal("expected error for missing coordinates")
	}
}

func TestResolveAndStoreCacheHit(t *testing.T) {
	if err := logger.Init("error", "text", ""); err != nil {
		t.Fatal(err)
	}
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
	_, err = db.Exec(
		`INSERT INTO locations
		 (raw_text, city, admin_area, country, latitude, longitude)
		 VALUES ('Portland', 'Portland', 'OR', 'US', 45.5152, -122.6784)`,
	)
	if err != nil {
		t.Fatal(err)
	}

	lat, lon, ok, err := ResolveLocation("Portland")
	if err != nil || !ok {
		t.Fatalf("ResolveLocation: ok=%v err=%v", ok, err)
	}
	if lat != 45.5152 || lon != -122.6784 {
		t.Fatalf("unexpected coords: %v, %v", lat, lon)
	}
}

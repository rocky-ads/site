package location

import (
	"errors"
	"os"
	"testing"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
)

type fakeLocationResolver struct {
	calls int
	resp  *LocationResponse
	err   error
}

func (f *fakeLocationResolver) Resolve(text string) (*LocationResponse, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}

func TestResolveAndStore(t *testing.T) {
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
	defer db.Close()
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}

	t.Run("cache miss inserts row", func(t *testing.T) {
		lat := 44.5646
		lon := -123.2620
		fake := &fakeLocationResolver{
			resp: &LocationResponse{
				City:      "Corvallis",
				AdminArea: "OR",
				Country:   "US",
				Latitude:  &lat,
				Longitude: &lon,
			},
		}
		SetLocationResolver(fake)
		t.Cleanup(func() { SetLocationResolver(nil) })

		id, ok, err := FindLocationID("97333")
		if err != nil || !ok {
			t.Fatalf("FindLocationID: ok=%v err=%v", ok, err)
		}
		if fake.calls != 1 {
			t.Fatalf("expected 1 resolver call, got %d", fake.calls)
		}

		var rawText, city, adminArea, country string
		var gotLat, gotLon float64
		err = db.QueryRow(
			`SELECT raw_text, city, admin_area, country, latitude, longitude
			 FROM locations WHERE id = ?`, id,
		).Scan(&rawText, &city, &adminArea, &country, &gotLat, &gotLon)
		if err != nil {
			t.Fatal(err)
		}
		if rawText != "97333" || city != "Corvallis" || adminArea != "OR" ||
			country != "US" {
			t.Fatalf("unexpected row: %q %q %q %q",
				rawText, city, adminArea, country)
		}

		resLat, resLon, ok, err := ResolveLocation("97333")
		if err != nil || !ok {
			t.Fatalf("ResolveLocation: ok=%v err=%v", ok, err)
		}
		if resLat != gotLat || resLon != gotLon {
			t.Fatalf("coords mismatch: got %v,%v want %v,%v",
				resLat, resLon, gotLat, gotLon)
		}
		if fake.calls != 1 {
			t.Fatalf("expected cache hit, resolver calls=%d", fake.calls)
		}
	})

	t.Run("cache hit skips resolver", func(t *testing.T) {
		_, err := db.Exec(
			`INSERT INTO locations
			 (raw_text, city, admin_area, country, latitude, longitude)
			 VALUES ('Portland', 'Portland', 'OR', 'US', 45.5152, -122.6784)`,
		)
		if err != nil {
			t.Fatal(err)
		}

		fake := &fakeLocationResolver{}
		SetLocationResolver(fake)
		t.Cleanup(func() { SetLocationResolver(nil) })

		lat, lon, ok, err := ResolveLocation("Portland")
		if err != nil || !ok {
			t.Fatalf("ResolveLocation: ok=%v err=%v", ok, err)
		}
		if lat != 45.5152 || lon != -122.6784 {
			t.Fatalf("unexpected coords: %v, %v", lat, lon)
		}
		if fake.calls != 0 {
			t.Fatalf("expected no resolver calls, got %d", fake.calls)
		}
	})

	t.Run("resolver failure is silent", func(t *testing.T) {
		fake := &fakeLocationResolver{err: errors.New("grok down")}
		SetLocationResolver(fake)
		t.Cleanup(func() { SetLocationResolver(nil) })

		id, ok, err := FindLocationID("99999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatalf("expected ok=false, got id=%d", id)
		}
	})
}

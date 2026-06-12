package location

import (
	"testing"
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

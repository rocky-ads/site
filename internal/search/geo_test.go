package search

import (
	"strings"
	"testing"
)

func TestGeoInAreaExpr(t *testing.T) {
	var pa pgArgs
	p := Params{
		CenterLat: 39.7392,
		CenterLon: -104.9903,
		WithinKm:  40.234,
		HasGeo:    true,
	}
	clause := geoInAreaExpr(p, &pa)
	if len(pa.args) != 4 {
		t.Fatalf("expected 4 bbox args, got %d", len(pa.args))
	}
	if !strings.Contains(clause, "latitude BETWEEN") {
		t.Fatalf("unexpected clause: %s", clause)
	}
	if strings.Contains(clause, "vector_metadata") {
		t.Fatalf("should not use metadata lat/lon: %s", clause)
	}
}

func TestGeoOutOfAreaExprNullSafe(t *testing.T) {
	var pa pgArgs
	p := Params{
		CategoryID: 1,
		HasGeo:     true,
		CenterLat:  45.5,
		CenterLon:  -122.6,
		WithinKm:   25,
	}
	bbox := geoInAreaExpr(p, &pa)
	expr := geoOutOfAreaExpr(bbox)
	if !strings.Contains(expr, "latitude IS NOT NULL") {
		t.Fatalf("expected null-safe out-of-area: %s", expr)
	}
	if strings.Contains(expr, "vector_metadata") {
		t.Fatalf("should not use metadata: %s", expr)
	}
}

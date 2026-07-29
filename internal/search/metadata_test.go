package search

import (
	"strings"
	"testing"
)

func TestGeoInAreaWhereClause(t *testing.T) {
	var pa pgArgs
	p := Params{
		CenterLat: 39.7392,
		CenterLon: -104.9903,
		WithinKm:  40.234,
		HasGeo:    true,
	}
	clause := geoInAreaWhereClause(p, &pa)
	if len(pa.args) != 4 {
		t.Fatalf("expected 4 bbox args, got %d", len(pa.args))
	}
	if !strings.Contains(clause, "vector_metadata->'location'->>'lat'") {
		t.Fatalf("unexpected clause: %s", clause)
	}
}

func TestBuildVectorMetadataWhereUsesRockColumn(t *testing.T) {
	var pa pgArgs
	p := Params{CategoryID: 6}
	clause := buildVectorMetadataWhere(p, &pa)
	if strings.Contains(clause, "rock_count')") {
		t.Fatalf("should not use metadata rock_count: %s", clause)
	}
	if !strings.Contains(clause, "rock_count <=") {
		t.Fatalf("expected rock_count column filter: %s", clause)
	}
}

func TestBuildVectorMetadataWhereNoGeoFilter(t *testing.T) {
	var pa pgArgs
	p := Params{
		CategoryID: 6,
		Expanded:   true,
		HasGeo:     true,
		CenterLat:  34.0,
		CenterLon:  -118.0,
		WithinKm:   50,
	}
	clause := buildVectorMetadataWhere(p, &pa)
	if strings.Contains(clause, "BETWEEN") {
		t.Fatalf("geo should not hard-filter results: %s", clause)
	}
}

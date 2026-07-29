package search

import (
	"strings"
	"testing"
)

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

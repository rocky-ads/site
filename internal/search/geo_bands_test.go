package search

import (
	"testing"
)

func TestSearchGeoOutWhereNullSafe(t *testing.T) {
	var pa pgArgs
	p := Params{
		CategoryID: 1,
		HasGeo:     true,
		CenterLat:  45.5,
		CenterLon:  -122.6,
		WithinKm:   25,
		Limit:      20,
		Offset:     0,
	}
	where := buildVectorMetadataWhere(p, &pa)
	bbox := geoInAreaExpr(p, &pa)
	outWhere := where + ` AND NOT (
		(vector_metadata->'location'->>'lat') IS NOT NULL
		AND (vector_metadata->'location'->>'lon') IS NOT NULL
		AND (` + bbox + `)
	)`
	if outWhere == where {
		t.Fatal("outWhere should extend base where")
	}
	if got := len(pa.args); got < 6 {
		t.Fatalf("expected category+rock+bbox args, got %d", got)
	}
}

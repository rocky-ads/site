package ad

import "testing"

func TestCategoryFacets(t *testing.T) {
	c := Category{FacetKeys: []string{"mileage", "price"}}
	defs := c.Facets()
	if len(defs) != 2 {
		t.Fatalf("expected 2 facet defs, got %d", len(defs))
	}
	if defs[0].Key != "mileage" || defs[1].Key != "price" {
		t.Fatalf("unexpected facet order: %+v", defs)
	}
}

func TestCategoryFacetsSkipsUnknown(t *testing.T) {
	c := Category{FacetKeys: []string{"mileage", "bogus"}}
	defs := c.Facets()
	if len(defs) != 1 || defs[0].Key != "mileage" {
		t.Fatalf("expected only mileage, got %+v", defs)
	}
}

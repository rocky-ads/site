package ad

import "testing"

func TestCategoryFacetFlags(t *testing.T) {
	tests := []struct {
		name       string
		flags      int
		hasMileage bool
		hasHours   bool
	}{
		{"mileage only", CategoryFlagMileage, true, false},
		{"hours only", CategoryFlagHours, false, true},
		{"both", CategoryFlagMileage | CategoryFlagHours, true, true},
		{"none", 0, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Category{Flags: tt.flags}
			if c.HasMileage() != tt.hasMileage {
				t.Fatalf("HasMileage() = %v, want %v", c.HasMileage(), tt.hasMileage)
			}
			if c.HasHours() != tt.hasHours {
				t.Fatalf("HasHours() = %v, want %v", c.HasHours(), tt.hasHours)
			}
		})
	}
}

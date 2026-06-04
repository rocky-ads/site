package location

import "testing"

func TestDistanceUnitFromPhone(t *testing.T) {
	tests := []struct {
		phone string
		want  string
	}{
		{"+14155552671", UnitMiles},
		{"+442079460123", UnitMiles},
		{"+33123456789", UnitKm},
		{"", UnitMiles},
	}
	for _, tt := range tests {
		if got := DistanceUnitFromPhone(tt.phone); got != tt.want {
			t.Errorf("DistanceUnitFromPhone(%q) = %q, want %q", tt.phone, got, tt.want)
		}
	}
}

func TestDistanceUnitFromTimezone(t *testing.T) {
	tests := []struct {
		tz   string
		want string
	}{
		{"America/Los_Angeles", UnitMiles},
		{"America/New_York", UnitMiles},
		{"Europe/London", UnitMiles},
		{"America/Toronto", UnitKm},
		{"Europe/Paris", UnitKm},
		{"", UnitKm},
	}
	for _, tt := range tests {
		if got := DistanceUnitFromTimezone(tt.tz); got != tt.want {
			t.Errorf("DistanceUnitFromTimezone(%q) = %q, want %q", tt.tz, got, tt.want)
		}
	}
}

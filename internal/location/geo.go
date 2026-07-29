package location

import (
	"strings"
)

// MilesToKm converts statute miles to kilometers.
func MilesToKm(miles float64) float64 {
	return miles * 1.609344
}

// NormalizeMileageUnit returns mi or km; invalid values default to mi.
func NormalizeMileageUnit(unit string) string {
	if strings.TrimSpace(unit) == UnitKm {
		return UnitKm
	}
	return UnitMiles
}

// ValidMileageUnit reports whether unit is mi or km.
func ValidMileageUnit(unit string) bool {
	unit = strings.TrimSpace(unit)
	return unit == UnitMiles || unit == UnitKm
}

// ResolveLocation looks up lat/lon for user-entered location text,
// resolving via Grok and caching when not already stored.
func ResolveLocation(text string) (lat, lon float64, ok bool, err error) {
	loc, ok, err := resolveAndStore(text)
	if err != nil || !ok {
		return 0, 0, ok, err
	}
	return loc.lat, loc.lon, true, nil
}

// FindLocation resolves text to a cached location row id and coordinates.
func FindLocation(text string) (id int, lat, lon float64, ok bool, err error) {
	loc, ok, err := resolveAndStore(text)
	if err != nil || !ok {
		return 0, 0, 0, ok, err
	}
	return loc.id, loc.lat, loc.lon, true, nil
}

// FindLocationID looks up a location row ID for user-entered text,
// resolving via Grok and caching when not already stored.
func FindLocationID(text string) (int, bool, error) {
	id, _, _, ok, err := FindLocation(text)
	return id, ok, err
}

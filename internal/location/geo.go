package location

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/rocky-ads/site/internal/db"
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

// ResolveLocation looks up lat/lon for user-entered location text against seeded locations.
func ResolveLocation(text string) (lat, lon float64, ok bool, err error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, 0, false, nil
	}

	// Exact raw_text (case-insensitive).
	err = db.QueryRow(
		`SELECT latitude, longitude FROM locations WHERE lower(raw_text) = lower(?) LIMIT 1`,
		text,
	).Scan(&lat, &lon)
	if err == nil {
		return lat, lon, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, fmt.Errorf("resolve location: %w", err)
	}

	like := "%" + escapeLike(text) + "%"
	err = db.QueryRow(
		`SELECT latitude, longitude FROM locations
		 WHERE lower(city) LIKE lower(?) OR lower(admin_area) LIKE lower(?) OR lower(raw_text) LIKE lower(?)
		 ORDER BY length(raw_text) ASC LIMIT 1`,
		like, like, like,
	).Scan(&lat, &lon)
	if err == nil {
		return lat, lon, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	return 0, 0, false, fmt.Errorf("resolve location: %w", err)
}

// FindLocationID looks up a location row ID for user-entered text.
func FindLocationID(text string) (int, bool, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, false, nil
	}

	var id int
	err := db.QueryRow(
		`SELECT id FROM locations WHERE lower(raw_text) = lower(?) LIMIT 1`,
		text,
	).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, fmt.Errorf("find location: %w", err)
	}

	like := "%" + escapeLike(text) + "%"
	err = db.QueryRow(
		`SELECT id FROM locations
		 WHERE lower(city) LIKE lower(?) OR lower(admin_area) LIKE lower(?) OR lower(raw_text) LIKE lower(?)
		 ORDER BY length(raw_text) ASC LIMIT 1`,
		like, like, like,
	).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("find location: %w", err)
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

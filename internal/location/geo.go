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

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

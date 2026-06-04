package search

import (
	"fmt"
	"math"
	"strings"

	"github.com/rocky-ads/site/internal/db"
)

func Search(p Params) ([]int, error) {
	query := `
		SELECT a.id
		FROM ads a`
	args := []any{p.CategoryID}

	if p.HasGeo {
		query += `
		INNER JOIN locations l ON a.location_id = l.id`
	}

	query += `
		WHERE a.category_id = ? AND a.deleted_at IS NULL`
	if p.PriceMin != nil {
		query += ` AND a.price >= ?`
		args = append(args, *p.PriceMin)
	}
	if p.PriceMax != nil {
		query += ` AND a.price <= ?`
		args = append(args, *p.PriceMax)
	}
	if p.HasTextQuery() {
		pattern := "%" + escapeLike(p.Q) + "%"
		query += ` AND (LOWER(a.title) LIKE LOWER(?) OR LOWER(a.description) LIKE LOWER(?))`
		args = append(args, pattern, pattern)
	}
	if p.HasGeo {
		minLat, maxLat, minLon, maxLon := geoBoundingBox(p.CenterLat, p.CenterLon, p.RadiusKm)
		query += ` AND l.latitude BETWEEN ? AND ? AND l.longitude BETWEEN ? AND ?`
		args = append(args, minLat, maxLat, minLon, maxLon)
	}

	query += ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, p.Limit, p.Offset)

	wrapped := fmt.Sprintf(
		`SELECT COALESCE(json_group_array(id), '[]') FROM (%s) AS limited_ads`,
		query,
	)

	var adIDs []int
	if err := db.QueryJSON(&adIDs, wrapped, args...); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return adIDs, nil
}

// geoBoundingBox returns an approximate lat/lon window for radiusKm (SQLite-safe).
func geoBoundingBox(lat, lon, radiusKm float64) (minLat, maxLat, minLon, maxLon float64) {
	const kmPerDegreeLat = 111.0
	deltaLat := radiusKm / kmPerDegreeLat
	cosLat := math.Cos(lat * math.Pi / 180)
	deltaLon := radiusKm / kmPerDegreeLat
	if cosLat > 0.01 {
		deltaLon = radiusKm / (kmPerDegreeLat * cosLat)
	}
	return lat - deltaLat, lat + deltaLat, lon - deltaLon, lon + deltaLon
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

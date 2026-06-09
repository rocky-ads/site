package search

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/facet"
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

	for _, key := range sortedFacetKeys(p.FacetFilters) {
		f := p.FacetFilters[key]
		if !f.Active() {
			continue
		}
		d, ok := facet.Get(key)
		if !ok {
			continue
		}
		switch d.Kind {
		case facet.MultiEnum:
			if len(f.Values) == 0 {
				continue
			}
			placeholders := make([]string, len(f.Values))
			facetArgs := []any{key}
			for i, v := range f.Values {
				placeholders[i] = "?"
				facetArgs = append(facetArgs, v)
			}
			clause := ` AND EXISTS (SELECT 1 FROM ad_facets f, json_each(f.text) je
				WHERE f.ad_id = a.id AND f.key = ? AND je.value IN (` +
				strings.Join(placeholders, ",") + `))`
			query += clause
			args = append(args, facetArgs...)
		case facet.Date:
			if f.TextMin == nil && f.TextMax == nil {
				continue
			}
			clause := ` AND EXISTS (SELECT 1 FROM ad_facets f
				WHERE f.ad_id = a.id AND f.key = ?`
			facetArgs := []any{key}
			if f.TextMin != nil {
				clause += ` AND f."text" >= ?`
				facetArgs = append(facetArgs, *f.TextMin)
			}
			if f.TextMax != nil {
				clause += ` AND f."text" <= ?`
				facetArgs = append(facetArgs, *f.TextMax)
			}
			clause += `)`
			query += clause
			args = append(args, facetArgs...)
		default:
			clause := ` AND EXISTS (SELECT 1 FROM ad_facets f
				WHERE f.ad_id = a.id AND f.key = ?`
			facetArgs := []any{key}
			switch {
			case len(f.Values) > 0:
				placeholders := make([]string, len(f.Values))
				for i, v := range f.Values {
					placeholders[i] = "?"
					facetArgs = append(facetArgs, v)
				}
				clause += ` AND f."text" IN (` + strings.Join(placeholders, ",") + `)`
			case f.Value != nil:
				clause += ` AND f."text" = ?`
				facetArgs = append(facetArgs, *f.Value)
			}
			if f.Min != nil {
				clause += ` AND f.num >= ?`
				facetArgs = append(facetArgs, *f.Min)
			}
			if f.Max != nil {
				clause += ` AND f.num <= ?`
				facetArgs = append(facetArgs, *f.Max)
			}
			clause += `)`
			query += clause
			args = append(args, facetArgs...)
		}
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

func sortedFacetKeys(filters map[string]facet.Filter) []string {
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

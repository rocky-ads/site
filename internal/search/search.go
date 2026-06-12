package search

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/facet"
)

type pgArgs struct {
	args []any
}

func (p *pgArgs) add(v any) string {
	p.args = append(p.args, v)
	return fmt.Sprintf("$%d", len(p.args))
}

func Search(p Params) ([]int, error) {
	var pa pgArgs
	query := `
		SELECT a.id
		FROM ads a`
	if p.HasGeo {
		query += `
		INNER JOIN locations l ON a.location_id = l.id`
	}

	query += `
		WHERE a.category_id = ` + pa.add(p.CategoryID) + ` AND a.deleted_at IS NULL`

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
			keyPH := pa.add(key)
			valuePH := make([]string, len(f.Values))
			for i, v := range f.Values {
				valuePH[i] = pa.add(v)
			}
			clause := ` AND EXISTS (SELECT 1 FROM ad_facets f, jsonb_array_elements_text(f.text::jsonb) AS je(value)
				WHERE f.ad_id = a.id AND f.key = ` + keyPH + ` AND je.value IN (` +
				strings.Join(valuePH, ",") + `))`
			query += clause
		case facet.Date:
			if f.TextMin == nil && f.TextMax == nil {
				continue
			}
			clause := ` AND EXISTS (SELECT 1 FROM ad_facets f
				WHERE f.ad_id = a.id AND f.key = ` + pa.add(key)
			if f.TextMin != nil {
				clause += ` AND f."text" >= ` + pa.add(*f.TextMin)
			}
			if f.TextMax != nil {
				clause += ` AND f."text" <= ` + pa.add(*f.TextMax)
			}
			clause += `)`
			query += clause
		default:
			clause := ` AND EXISTS (SELECT 1 FROM ad_facets f
				WHERE f.ad_id = a.id AND f.key = ` + pa.add(key)
			switch {
			case len(f.Values) > 0:
				valuePH := make([]string, len(f.Values))
				for i, v := range f.Values {
					valuePH[i] = pa.add(v)
				}
				clause += ` AND f."text" IN (` + strings.Join(valuePH, ",") + `)`
			case f.Value != nil:
				clause += ` AND f."text" = ` + pa.add(*f.Value)
			}
			if f.Min != nil {
				clause += ` AND f.num >= ` + pa.add(*f.Min)
			}
			if f.Max != nil {
				clause += ` AND f.num <= ` + pa.add(*f.Max)
			}
			clause += `)`
			query += clause
		}
	}

	if p.HasTextQuery() {
		pattern := "%" + escapeLike(p.Q) + "%"
		titlePH := pa.add(pattern)
		descPH := pa.add(pattern)
		query += ` AND (a.title ILIKE ` + titlePH + ` OR a.description ILIKE ` + descPH + `)`
	}
	if p.HasGeo {
		minLat, maxLat, minLon, maxLon := geoBoundingBox(p.CenterLat, p.CenterLon, p.RadiusKm)
		minLatPH := pa.add(minLat)
		maxLatPH := pa.add(maxLat)
		minLonPH := pa.add(minLon)
		maxLonPH := pa.add(maxLon)
		query += ` AND l.latitude BETWEEN ` + minLatPH + ` AND ` + maxLatPH +
			` AND l.longitude BETWEEN ` + minLonPH + ` AND ` + maxLonPH
	}

	limitPH := pa.add(p.Limit)
	offsetPH := pa.add(p.Offset)
	query += ` ORDER BY a.created_at DESC LIMIT ` + limitPH + ` OFFSET ` + offsetPH

	wrapped := fmt.Sprintf(
		`SELECT COALESCE(json_agg(id), '[]'::json) FROM (%s) AS limited_ads`,
		query,
	)

	var adIDs []int
	if err := db.QueryJSON(&adIDs, wrapped, pa.args...); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return adIDs, nil
}

// geoBoundingBox returns an approximate lat/lon window for radiusKm.
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

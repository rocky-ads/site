package search

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/facet"
)

func buildVectorMetadataWhere(p Params, pa *pgArgs) string {
	clause := `(vector_metadata->>'category_id')::int = ` +
		pa.add(p.CategoryID)
	clause += ` AND COALESCE((vector_metadata->>'egg_count')::int, 0) <= ` +
		pa.add(config.MaxEggCount)

	if !p.Expanded {
		return clause
	}

	for _, key := range sortedFacetKeys(p.FacetFilters) {
		f := p.FacetFilters[key]
		if !f.Active() {
			continue
		}
		d, ok := facet.Get(key)
		if !ok {
			continue
		}
		part := vectorFilterClause(d, f, pa)
		if part != "" {
			clause += " AND " + part
		}
	}

	return clause
}

func vectorFilterClause(d facet.Def, f facet.Filter, pa *pgArgs) string {
	path := `vector_metadata->>'` + d.Key + `'`
	switch d.Kind {
	case facet.MultiEnum:
		if len(f.Values) == 0 {
			return ""
		}
		valuePH := make([]string, len(f.Values))
		for i, v := range f.Values {
			valuePH[i] = pa.add(v)
		}
		return fmt.Sprintf(
			`EXISTS (SELECT 1 FROM jsonb_array_elements_text(`+
				`vector_metadata->'%s') AS je(value) `+
				`WHERE je.value IN (%s))`,
			d.Key, strings.Join(valuePH, ","),
		)
	case facet.Date:
		clause := ""
		if f.TextMin != nil {
			clause += path + ` >= ` + pa.add(*f.TextMin)
		}
		if f.TextMax != nil {
			if clause != "" {
				clause += " AND "
			}
			clause += path + ` <= ` + pa.add(*f.TextMax)
		}
		return clause
	case facet.Money, facet.Int:
		numPath := `(` + path + `)::int`
		clause := ""
		if f.Min != nil {
			clause += numPath + ` >= ` + pa.add(*f.Min)
		}
		if f.Max != nil {
			if clause != "" {
				clause += " AND "
			}
			clause += numPath + ` <= ` + pa.add(*f.Max)
		}
		return clause
	default:
		clause := ""
		if len(f.Values) > 0 {
			valuePH := make([]string, len(f.Values))
			for i, v := range f.Values {
				valuePH[i] = pa.add(v)
			}
			clause = path + ` IN (` + strings.Join(valuePH, ",") + `)`
		} else if f.Value != nil {
			clause = path + ` = ` + pa.add(*f.Value)
		}
		if f.Min != nil {
			numPath := `(` + path + `)::int`
			if clause != "" {
				clause += " AND "
			}
			clause += numPath + ` >= ` + pa.add(*f.Min)
		}
		if f.Max != nil {
			numPath := `(` + path + `)::int`
			if clause != "" {
				clause += " AND "
			}
			clause += numPath + ` <= ` + pa.add(*f.Max)
		}
		return clause
	}
}

func sortedFacetKeys(filters map[string]facet.Filter) []string {
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func geoBoundingBox(
	lat, lon, withinKm float64,
) (minLat, maxLat, minLon, maxLon float64) {
	const kmPerDegreeLat = 111.0
	deltaLat := withinKm / kmPerDegreeLat
	cosLat := math.Cos(lat * math.Pi / 180)
	deltaLon := withinKm / kmPerDegreeLat
	if cosLat > 0.01 {
		deltaLon = withinKm / (kmPerDegreeLat * cosLat)
	}
	return lat - deltaLat, lat + deltaLat, lon - deltaLon, lon + deltaLon
}

const (
	latMetaKey = `(vector_metadata->'location'->>'lat')::float`
	lonMetaKey = `(vector_metadata->'location'->>'lon')::float`
)

func geoInAreaWhereClause(p Params, pa *pgArgs) string {
	minLat, maxLat, minLon, maxLon := geoBoundingBox(
		p.CenterLat, p.CenterLon, p.WithinKm,
	)
	return latMetaKey + ` BETWEEN ` + pa.add(minLat) +
		` AND ` + pa.add(maxLat) +
		` AND ` + lonMetaKey + ` BETWEEN ` + pa.add(minLon) +
		` AND ` + pa.add(maxLon)
}

func geoInAreaOrderExpr(p Params, pa *pgArgs) string {
	minLat, maxLat, minLon, maxLon := geoBoundingBox(
		p.CenterLat, p.CenterLon, p.WithinKm,
	)
	return `CASE WHEN ` + latMetaKey + ` BETWEEN ` + pa.add(minLat) +
		` AND ` + pa.add(maxLat) +
		` AND ` + lonMetaKey + ` BETWEEN ` + pa.add(minLon) +
		` AND ` + pa.add(maxLon) +
		` THEN 0 ELSE 1 END`
}

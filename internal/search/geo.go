package search

import "math"

func geoBoundingBox(lat, lon,
	withinKm float64) (minLat, maxLat, minLon, maxLon float64) {
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
	latCol = `latitude`
	lonCol = `longitude`
)

func geoInAreaExpr(p Params, pa *pgArgs) string {
	minLat, maxLat, minLon, maxLon := geoBoundingBox(
		p.CenterLat, p.CenterLon, p.WithinKm,
	)
	return latCol + ` BETWEEN ` + pa.add(minLat) +
		` AND ` + pa.add(maxLat) +
		` AND ` + lonCol + ` BETWEEN ` + pa.add(minLon) +
		` AND ` + pa.add(maxLon)
}

func geoOutOfAreaExpr(bbox string) string {
	return `NOT (
		` + latCol + ` IS NOT NULL
		AND ` + lonCol + ` IS NOT NULL
		AND (` + bbox + `)
	)`
}

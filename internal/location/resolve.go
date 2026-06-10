package location

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/service/grok"
)

const locationResolverConvID = "location-resolver"

const locationResolverPrompt = `You are a location resolver for classified ads
website.  Given a user input (which may be a address, city, state, zip code, or
country), return a JSON object with the best guess for city, admin_area (state,
province, or region), country, latitude, and longitude. The country field must
be a 2-letter ISO country code (e.g., "US" for United States, "CA" for Canada,
"GB" for United Kingdom). For US and Canada, the admin_area field must be the
official 2-letter code (e.g., "OR" for Oregon, "NY" for New York, "BC" for
British Columbia, "ON" for Ontario). For all other countries, use the full name
for admin_area. Latitude and longitude should be decimal degrees (positive for
North/East, negative for South/West).  If a field is unknown, leave it blank or
null.  Example input: "97333" -> {"city": "Corvallis", "admin_area": "OR",
"country": "US", "latitude": 44.5646, "longitude": -123.2620}`

// LocationResponse is the JSON shape returned by Grok location resolution.
type LocationResponse struct {
	City      string   `json:"city"`
	AdminArea string   `json:"admin_area"`
	Country   string   `json:"country"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func resolveWithGrok(text string) (*LocationResponse, error) {
	resp, err := grok.CallGrokConv(
		locationResolverPrompt, text, locationResolverConvID,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve location with grok: %w", err)
	}
	return parseLocationResponse(resp)
}

type resolvedLoc struct {
	id       int
	lat, lon float64
}

func resolveAndStore(text string) (resolvedLoc, bool, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return resolvedLoc{}, false, nil
	}

	var loc resolvedLoc
	err := db.QueryRow(
		`SELECT id, latitude, longitude FROM locations
		 WHERE lower(raw_text) = lower(?) LIMIT 1`,
		text,
	).Scan(&loc.id, &loc.lat, &loc.lon)
	if err == nil {
		return loc, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return resolvedLoc{}, false, fmt.Errorf("resolve location: %w", err)
	}

	resolved, err := resolveWithGrok(text)
	if err != nil {
		logger.Warn("resolve location: grok failed", "text", text, "error", err)
		return resolvedLoc{}, false, nil
	}
	if resolved.Latitude == nil || resolved.Longitude == nil {
		logger.Warn("resolve location: missing coordinates", "text", text)
		return resolvedLoc{}, false, nil
	}

	_, err = db.Exec(
		`INSERT INTO locations
		 (raw_text, city, admin_area, country, latitude, longitude)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(raw_text) DO NOTHING`,
		text, resolved.City, resolved.AdminArea, resolved.Country,
		*resolved.Latitude, *resolved.Longitude,
	)
	if err != nil {
		return resolvedLoc{}, false, fmt.Errorf("resolve location: %w", err)
	}

	err = db.QueryRow(
		`SELECT id, latitude, longitude FROM locations
		 WHERE lower(raw_text) = lower(?) LIMIT 1`,
		text,
	).Scan(&loc.id, &loc.lat, &loc.lon)
	if err != nil {
		return resolvedLoc{}, false, fmt.Errorf("resolve location: %w", err)
	}
	return loc, true, nil
}

func parseLocationResponse(resp string) (*LocationResponse, error) {
	resp = strings.TrimSpace(trimCodeFence(resp))
	var loc LocationResponse
	if err := json.Unmarshal([]byte(resp), &loc); err != nil {
		return nil, fmt.Errorf("parse location response: %w", err)
	}
	if loc.Latitude == nil || loc.Longitude == nil {
		return nil, fmt.Errorf("location response missing coordinates")
	}
	return &loc, nil
}

func trimCodeFence(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return s
	}
	end := len(lines) - 1
	for end > 0 && strings.TrimSpace(lines[end]) == "" {
		end--
	}
	if strings.HasPrefix(strings.TrimSpace(lines[end]), "```") {
		return strings.Join(lines[1:end], "\n")
	}
	return strings.Join(lines[1:], "\n")
}

package ad

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/location"
)

type CreateInput struct {
	CategoryID   int
	UserID       int
	Title        string
	Description  string
	LocationText string
	Facets       map[string]facet.Value
	Suggestions  []Suggestion
	ImageCount   int
}

func CreateAd(input CreateInput) (int, error) {
	category := GetCategory(input.CategoryID)

	title := strings.TrimSpace(SanitizeAdText(input.Title))
	description := strings.TrimSpace(SanitizeAdText(input.Description))
	if title == "" {
		return 0, fmt.Errorf("title is required")
	}
	if description == "" {
		return 0, fmt.Errorf("description is required")
	}
	if utf8.RuneCountInString(title) > config.MaxAdTitleLength {
		return 0, fmt.Errorf("title must be at most %d characters", config.MaxAdTitleLength)
	}
	if TitleContainsEmoji(title) {
		return 0, fmt.Errorf("title cannot contain emoji")
	}
	if utf8.RuneCountInString(description) > config.MaxAdDescriptionLength {
		return 0, fmt.Errorf("description must be at most %d characters", config.MaxAdDescriptionLength)
	}
	description = WrapDescription(description, time.Now().UTC(), time.UTC)
	defs := category.Facets()
	values := make(map[string]facet.Value, len(defs))
	for _, d := range defs {
		v := input.Facets[d.Key]
		if err := d.Validate(v); err != nil {
			return 0, err
		}
		if v.Present() {
			values[d.Key] = v
		}
	}

	locText := input.LocationText
	if HasLocationFacet(category) {
		locText = LocationTextFromFacets(category, values)
	}
	var locationID any
	if strings.TrimSpace(locText) != "" {
		id, ok, err := location.FindLocationID(locText)
		if err != nil {
			return 0, err
		}
		if ok {
			locationID = id
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("create ad: %w", err)
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRow(
		`INSERT INTO ads (category_id, title, description, user_id, location_id,
		 tags, image_count)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		input.CategoryID, title, description, input.UserID, locationID,
		tagsJSON(input.Suggestions), input.ImageCount,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create ad: %w", err)
	}

	for key, v := range values {
		if _, err := tx.Exec(
			`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES ($1, $2, $3, $4)`,
			id, key, v.Num, v.Text,
		); err != nil {
			return 0, fmt.Errorf("create ad facet %s: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("create ad commit: %w", err)
	}
	return id, nil
}

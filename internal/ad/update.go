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

type UpdateInput struct {
	AdID                int
	UserID              int
	Title               string
	DescriptionAddition string
	LocationText        string
	Facets              map[string]facet.Value
	Suggestions         []Suggestion
	ImagesAdded         int
	Loc                 *time.Location
	Now                 time.Time
}

func UpdateAd(input UpdateInput) error {
	a, err := GetAd(input.UserID, input.AdID, input.Loc)
	if err != nil {
		return err
	}
	if a.UserID != input.UserID {
		return fmt.Errorf("you are not the owner of this ad")
	}
	if a.IsDeleted() {
		return fmt.Errorf("cannot edit a deleted ad")
	}

	category, err := GetCategory(a.CategoryID)
	if err != nil {
		return err
	}

	title := strings.TrimSpace(SanitizeAdText(input.Title))
	if title == "" {
		return fmt.Errorf("title is required")
	}
	if utf8.RuneCountInString(title) > config.MaxAdTitleLength {
		return fmt.Errorf(
			"title must be at most %d characters",
			config.MaxAdTitleLength,
		)
	}
	if TitleContainsEmoji(title) {
		return fmt.Errorf("title cannot contain emoji")
	}

	addition := strings.TrimSpace(SanitizeAdText(input.DescriptionAddition))
	if strings.Contains(addition, historyMarker) ||
		strings.Contains(addition, historyEndMarker) {
		return fmt.Errorf("invalid description addition")
	}

	if err := LoadTags(&a); err != nil {
		return err
	}

	newSuggestions := dedupeSuggestions(input.Suggestions)
	defs := category.Facets()
	values := make(map[string]facet.Value, len(defs))
	for _, d := range defs {
		v := input.Facets[d.Key]
		if err := d.Validate(v); err != nil {
			return err
		}
		if v.Present() {
			values[d.Key] = v
		}
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}

	desc := a.Description
	if addition != "" {
		desc = AppendHistoryEntry(
			desc, "Description Addition", addition, now, input.Loc,
		)
	}
	if input.ImagesAdded > 0 {
		if a.ImageCount+input.ImagesAdded > config.MaxImagesPerAd {
			return fmt.Errorf(
				"too many images. Maximum %d images allowed per ad",
				config.MaxImagesPerAd,
			)
		}
		body := formatImageAdditionBody(a.ImageCount+1, input.ImagesAdded)
		desc = AppendHistoryEntry(
			desc, imagesAddedLabel, body, now, input.Loc,
		)
	}
	locText := input.LocationText
	if HasLocationFacet(category) {
		locText = LocationTextFromFacets(category, values)
	}
	for _, e := range BuildFieldChangeEntries(
		a, title, locText, values, category,
	) {
		desc = AppendHistoryEntry(desc, e.label, e.body, now, input.Loc)
	}
	if body := FormatTagUpdates(a.Tags, newSuggestions); body != "" {
		desc = AppendHistoryEntry(
			desc, "Description Tags", body, now, input.Loc,
		)
	}

	desc, err = EnsureDescriptionFits(desc, now, input.Loc)
	if err != nil {
		return err
	}

	var locationID any
	if strings.TrimSpace(locText) != "" {
		id, ok, err := location.FindLocationID(locText)
		if err != nil {
			return err
		}
		if ok {
			locationID = id
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("update ad: %w", err)
	}
	defer tx.Rollback()

	newImageCount := a.ImageCount + input.ImagesAdded

	_, err = tx.Exec(
		`UPDATE ads SET title = ?, description = ?, location_id = ?, tags = ?,
		 image_count = ?
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		title, desc, locationID, tagsJSON(newSuggestions), newImageCount,
		input.AdID, input.UserID,
	)
	if err != nil {
		return fmt.Errorf("update ad: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM ad_facets WHERE ad_id = ?`, input.AdID); err != nil {
		return fmt.Errorf("update ad facets: %w", err)
	}
	for key, v := range values {
		if _, err := tx.Exec(
			`INSERT INTO ad_facets (ad_id, "key", num, "text") VALUES (?, ?, ?, ?)`,
			input.AdID, key, v.Num, v.Text,
		); err != nil {
			return fmt.Errorf("update ad facet %s: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update ad commit: %w", err)
	}
	return nil
}

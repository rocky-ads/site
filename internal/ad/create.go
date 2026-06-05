package ad

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/location"
)

var (
	printableASCII      = regexp.MustCompile(`^[\x20-\x7E]+$`)
	printableASCIIMulti = regexp.MustCompile(`^[\x20-\x7E\n\r]+$`)
)

type CreateInput struct {
	CategoryID    int
	UserID        int
	Title         string
	Description   string
	Price         int
	PriceCurrency string
	LocationText  string
	Mileage       *int
	MileageUnit   *string
	Hours         *int
}

func CreateAd(input CreateInput) (int, error) {
	category, err := GetCategory(input.CategoryID)
	if err != nil {
		return 0, err
	}

	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	if title == "" {
		return 0, fmt.Errorf("title is required")
	}
	if description == "" {
		return 0, fmt.Errorf("description is required")
	}
	if utf8.RuneCountInString(title) > config.MaxAdTitleLength {
		return 0, fmt.Errorf("title must be at most %d characters", config.MaxAdTitleLength)
	}
	if utf8.RuneCountInString(description) > config.MaxAdDescriptionLength {
		return 0, fmt.Errorf("description must be at most %d characters", config.MaxAdDescriptionLength)
	}
	if !printableASCII.MatchString(title) {
		return 0, fmt.Errorf("title must contain printable ASCII characters only")
	}
	if !printableASCIIMulti.MatchString(description) {
		return 0, fmt.Errorf("description must contain printable ASCII characters only")
	}
	if input.Price < 0 {
		return 0, fmt.Errorf("price must be zero or greater")
	}

	priceCurrency := currency.Normalize(input.PriceCurrency)
	if !currency.IsSupported(priceCurrency) {
		priceCurrency = currency.Default
	}

	var mileage *int
	var mileageUnit *string
	if category.HasMileage() {
		if input.Mileage != nil && *input.Mileage < 0 {
			return 0, fmt.Errorf("mileage must be zero or greater")
		}
		mileage = input.Mileage
		if mileage != nil {
			if input.MileageUnit == nil || !location.ValidMileageUnit(*input.MileageUnit) {
				return 0, fmt.Errorf("mileage unit must be mi or km")
			}
			unit := location.NormalizeMileageUnit(*input.MileageUnit)
			mileageUnit = &unit
		}
	}

	var hours *int
	if category.HasHours() {
		if input.Hours != nil && *input.Hours < 0 {
			return 0, fmt.Errorf("hours must be zero or greater")
		}
		hours = input.Hours
	}

	var locationID any
	if strings.TrimSpace(input.LocationText) != "" {
		id, ok, err := location.FindLocationID(input.LocationText)
		if err != nil {
			return 0, err
		}
		if ok {
			locationID = id
		}
	}

	result, err := db.Exec(
		`INSERT INTO ads (category_id, title, description, price, price_currency, user_id, location_id, mileage, mileage_unit, hours)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.CategoryID, title, description, input.Price, priceCurrency, input.UserID, locationID, mileage, mileageUnit, hours,
	)
	if err != nil {
		return 0, fmt.Errorf("create ad: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create ad id: %w", err)
	}
	return int(id), nil
}

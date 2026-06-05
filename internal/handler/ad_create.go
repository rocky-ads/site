package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/location"
	"github.com/rocky-ads/site/internal/ui"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
)

func CreateAdHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	if userID == 0 {
		return redirectToLogin(c)
	}

	categoryID := cookie.GetCategoryID(c)
	category, err := ad.GetCategory(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	price := 0
	if c.FormValue("price_free") != "1" {
		priceStr := strings.TrimSpace(c.FormValue("price"))
		if priceStr != "" {
			parsed, err := strconv.Atoi(priceStr)
			if err != nil || parsed < 0 {
				return showCreateAdError(c, category, "Price must be a non-negative whole number")
			}
			price = parsed
		}
	}

	var mileage *int
	var mileageUnit *string
	if category.HasMileage() {
		mileage, err = parseOptionalFacet(c.FormValue("mileage"))
		if err != nil {
			return showCreateAdError(c, category, "Mileage must be a non-negative whole number")
		}
		if mileage != nil {
			rawUnit := strings.TrimSpace(c.FormValue("mileage_unit"))
			if rawUnit == "" {
				rawUnit = distanceUnit(c)
			}
			if !location.ValidMileageUnit(rawUnit) {
				return showCreateAdError(c, category, "Mileage unit must be miles or km")
			}
			unit := location.NormalizeMileageUnit(rawUnit)
			mileageUnit = &unit
		}
	}

	var hours *int
	if category.HasHours() {
		hours, err = parseOptionalFacet(c.FormValue("hours"))
		if err != nil {
			return showCreateAdError(c, category, "Hours must be a non-negative whole number")
		}
	}

	adID, err := ad.CreateAd(ad.CreateInput{
		CategoryID:    categoryID,
		UserID:        userID,
		Title:         c.FormValue("title"),
		Description:   c.FormValue("description"),
		Price:         price,
		PriceCurrency: c.FormValue("price_currency"),
		LocationText:  c.FormValue("location"),
		Mileage:       mileage,
		MileageUnit:   mileageUnit,
		Hours:         hours,
	})
	if err != nil {
		return showCreateAdError(c, category, err.Error())
	}

	return c.Redirect("/ad/"+strconv.Itoa(adID), fiber.StatusFound)
}

func parseOptionalFacet(raw string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return nil, err
	}
	return &value, nil
}

func showCreateAdError(c *fiber.Ctx, category ad.Category, errMsg string) error {
	fieldsNode := uiads.NewAdFieldsPartial(category, defaultCurrencyForUser(c), distanceUnit(c))
	return renderPage(c, "New Ad", append(ui.NewAd(category.Name, fieldsNode), ui.ErrorDiv(errMsg)))
}

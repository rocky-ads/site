package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/eggopinion"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/param"
	"github.com/rocky-ads/site/internal/ui"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
	"github.com/rocky-ads/site/internal/vector"
)

func EditAdHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)

	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	loc := cookie.GetLocation(c)
	a, err := ad.GetAd(userID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}
	if a.IsDeleted() {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot edit a deleted ad")
	}

	category, err := ad.GetCategory(a.CategoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if err := ad.LoadTags(&a); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	values := adFormValuesFrom(a)
	cfg := uiads.EditFormConfig(adID, values, newAdFormDefaults(c))
	fieldsNode := uiads.AdFieldsPartial(cfg, category.Facets())

	return renderPage(c, "Edit Ad", ui.EditAd(categoryOption(category), cfg, fieldsNode))
}

func UpdateAdHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)

	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	loc := cookie.GetLocation(c)
	a, err := ad.GetAd(userID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}
	if a.IsDeleted() {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot edit a deleted ad")
	}

	category, err := ad.GetCategory(a.CategoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	facets, err := parseAdFacets(c, category)
	if err != nil {
		return showError(c, err.Error())
	}

	imageFiles, err := parseAdImageFiles(c)
	if err != nil {
		return showError(c, err.Error())
	}
	if err := validateAppendImageFiles(a.ImageCount, imageFiles); err != nil {
		return showError(c, err.Error())
	}

	err = ad.UpdateAd(ad.UpdateInput{
		AdID:                adID,
		UserID:              userID,
		Title:               c.FormValue("title"),
		DescriptionAddition: c.FormValue("description_addition"),
		LocationText:        c.FormValue("location"),
		Facets:              facets,
		Suggestions:         parseAdSuggestions(c),
		ImagesAdded:         len(imageFiles),
		Loc:                 loc,
	})
	if err != nil {
		return showError(c, err.Error())
	}
	if len(imageFiles) > 0 {
		uploadAdImagesFromIndex(
			adImageStore, adID, a.ImageCount+1, imageFiles,
		)
	}
	if err := eggopinion.InvalidateForAd(adID); err != nil {
		logger.Error("Failed to invalidate egg opinions",
			"error", err, "adID", adID)
	}
	vector.QueueAd(adID)

	redirect := "/ad/" + strconv.Itoa(adID)
	if c.Get("HX-Request") != "" {
		c.Set("HX-Redirect", redirect)
		return c.SendStatus(fiber.StatusOK)
	}
	return c.Redirect(redirect, fiber.StatusFound)
}

func adFormValuesFrom(a ad.Ad) uiads.AdFormValues {
	original, _ := ad.SplitDescription(a.Description)
	values := uiads.AdFormValues{
		Title:               a.Title,
		OriginalDescription: ad.DisplayDescription(original),
		Location:            a.RawLocation,
		ImageCount:          a.ImageCount,
		PriceRow:            priceRowFromAd(a),
		Facets:              make(map[string]string),
		FacetUnits:          make(map[string]string),
		FacetMulti:          make(map[string][]string),
	}
	for key, v := range a.Facets {
		if key == "price" {
			continue
		}
		d, ok := facet.Get(key)
		if ok && d.Kind == facet.MultiEnum {
			values.FacetMulti[key] = v.MultiEnumValues()
			continue
		}
		if v.Num != nil {
			values.Facets[key] = strconv.Itoa(*v.Num)
		}
		if v.Text != nil {
			if v.Num != nil {
				values.FacetUnits[key] = *v.Text
			} else {
				values.Facets[key] = *v.Text
			}
		}
	}
	for _, s := range a.Tags {
		values.Suggestions = append(values.Suggestions, uiads.SuggestionOption{
			Label:    s.Label,
			Value:    s.Value,
			Selected: true,
		})
	}
	return values
}

func priceRowFromAd(a ad.Ad) uiads.PriceRowView {
	v, ok := a.Facets["price"]
	if !ok || v.Num == nil {
		return uiads.PriceRowView{}
	}
	currency := ""
	if v.Text != nil {
		currency = *v.Text
	}
	if *v.Num == 0 {
		return uiads.PriceRowView{IsFree: true, Currency: currency}
	}
	return uiads.PriceRowView{
		Amount:   strconv.Itoa(*v.Num),
		Currency: currency,
	}
}

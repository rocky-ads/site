package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/local"
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

	facets, err := parseCreateFacets(c, category)
	if err != nil {
		return showCreateAdError(c, category, err.Error())
	}

	adID, err := ad.CreateAd(ad.CreateInput{
		CategoryID:   categoryID,
		UserID:       userID,
		Title:        c.FormValue("title"),
		Description:  c.FormValue("description"),
		LocationText: c.FormValue("location"),
		Facets:       facets,
	})
	if err != nil {
		return showCreateAdError(c, category, err.Error())
	}

	return c.Redirect("/ad/"+strconv.Itoa(adID), fiber.StatusFound)
}

func parseCreateFacets(c *fiber.Ctx, category ad.Category) (map[string]facet.Value, error) {
	values := make(map[string]facet.Value)
	for _, d := range category.Facets() {
		switch d.Kind {
		case facet.Money:
			free := c.FormValue("price_free") == "1"
			raw := strings.TrimSpace(c.FormValue(d.Key))
			code := currency.Normalize(c.FormValue("price_currency"))
			if !currency.IsSupported(code) {
				code = defaultCurrencyForUser(c)
			}
			if free {
				amount := 0
				values[d.Key] = facet.Value{Num: &amount, Text: &code}
				continue
			}
			if raw == "" {
				if d.Required {
					return nil, fmt.Errorf("%s is required", d.Label)
				}
				continue
			}
			n, err := strconv.Atoi(raw)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("%s must be a non-negative whole number", d.Label)
			}
			values[d.Key] = facet.Value{Num: &n, Text: &code}
		case facet.Enum:
			val := strings.TrimSpace(c.FormValue(d.Key))
			if val != "" {
				values[d.Key] = facet.Value{Text: &val}
			}
		default:
			num, err := parseOptionalFacet(c.FormValue(d.Key))
			if err != nil {
				return nil, fmt.Errorf("%s must be a non-negative whole number", d.Label)
			}
			if num == nil {
				continue
			}
			v := facet.Value{Num: num}
			if len(d.Units) > 0 {
				unit := strings.TrimSpace(c.FormValue(d.Key + "_unit"))
				if unit == "" {
					unit = distanceUnit(c)
				}
				if !d.ValidUnit(unit) {
					return nil, fmt.Errorf("%s requires a valid unit", d.Label)
				}
				u := d.NormalizeUnit(unit)
				v.Text = &u
			}
			values[d.Key] = v
		}
	}
	return values, nil
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
	fieldsNode := uiads.NewAdFieldsPartial(category.Facets(), newAdFormDefaults(c))
	return renderPage(c, "New Ad", append(ui.NewAd(categoryOption(category), fieldsNode), ui.ErrorDiv(errMsg)))
}

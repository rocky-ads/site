package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/param"
	"github.com/rocky-ads/site/internal/ui"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
	"github.com/rocky-ads/site/internal/user"
)

func NewAdHandler(c *fiber.Ctx) error {
	category := ad.GetCategory(cookie.GetCategoryID(c))
	fieldsNode := uiads.NewAdFieldsPartial(category.Facets(), newAdFormDefaults(c))

	return renderPage(c, "New Ad",
		ui.NewAd(categoryOption(category), allCategoryOptions(),
			fieldsNode))
}

func CopyAdHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)

	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	tz := cookie.GetTimezone(c)
	a, err := ad.GetAd(userID, adID, tz)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}
	if !a.IsActive() {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot copy a deleted or inactive ad")
	}

	cookie.SetCategoryID(c, a.CategoryID)
	category := ad.GetCategory(a.CategoryID)

	if err := ad.LoadTags(&a); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	values := adFormValuesFrom(a)
	cfg := uiads.NewFormConfigWithValues(newAdFormDefaults(c), values)
	fieldsNode := uiads.AdFieldsPartial(cfg, category.Facets())

	return renderPage(c, "New Ad",
		ui.NewAd(categoryOption(category), allCategoryOptions(),
			fieldsNode))
}

func NewAdPriceFieldHandler(c *fiber.Ctx) error {
	d, ok := facet.Get("price")
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "Price facet not found")
	}

	defaults := newAdFormDefaults(c)
	isFree := c.Query("price_free") == "1"
	amount := strings.TrimSpace(c.Query("price"))
	if isFree {
		amount = ""
	} else if amount == "0" {
		amount = ""
	}

	return render(c, uiads.NewAdPriceRow(d, defaults, uiads.PriceRowView{
		IsFree:   isFree,
		Amount:   amount,
		Currency: c.Query("price_currency"),
	}))
}

func newAdFormDefaults(c *fiber.Ctx) facet.FormDefaults {
	return facet.FormDefaults{
		Currency: defaultCurrencyForUser(c),
		Unit:     cookie.GetDistanceUnit(c),
	}
}

func defaultCurrencyForUser(c *fiber.Ctx) string {
	userID := local.GetUserID(c)
	if !local.IsLoggedIn(userID) {
		return currency.Default
	}
	u, err := user.GetByID(userID)
	if err != nil {
		return currency.Default
	}
	return currency.DefaultFromPhone(u.PhoneE64)
}

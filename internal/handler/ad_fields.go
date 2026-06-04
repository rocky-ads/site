package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/ui"
	uiads "github.com/rocky-ads/site/internal/ui/ads"
	"github.com/rocky-ads/site/internal/user"
)

func NewAdHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	if userID == 0 {
		return redirectToLogin(c)
	}

	categoryID := cookie.GetCategoryID(c)
	categoryName, err := ad.GetCategoryName(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	fieldsNode := uiads.NewAdFieldsPartial(defaultCurrencyForUser(c))

	return renderPage(c, "New Ad", ui.NewAd(categoryName, fieldsNode))
}

func defaultCurrencyForUser(c *fiber.Ctx) string {
	userID := local.GetUserID(c)
	if userID == 0 {
		return currency.Default
	}
	u, err := user.GetByID(userID)
	if err != nil {
		return currency.Default
	}
	return currency.DefaultFromPhone(u.PhoneE64)
}

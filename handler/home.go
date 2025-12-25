package handler

import (
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/ui"

	"github.com/gofiber/fiber/v2"
)

func HomeHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)

	categoryName, err := ad.GetCategoryName(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	categoryImage, err := ad.GetCategoryImageFile(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	userID := local.GetUserID(c)
	view := cookie.GetView(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)
	limit := config.SearchPageSize
	offset := 0

	results, err := searchAndRenderAds(categoryID, limit, offset, userID, view, make(field.Values), loc, csrfToken)
	if err != nil {
		return err
	}

	return renderPage(c, config.ServerName,
		ui.HomePage(userID, view, categoryName, categoryImage, results))
}

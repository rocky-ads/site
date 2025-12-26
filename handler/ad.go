package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/ui"
)

func AdHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	userID := local.GetUserID(c)
	loc := cookie.GetLocation(c)

	a, err := ad.GetAd(userID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	// If ad is deleted and user is not the owner, show deleted message
	if a.IsDeleted() && a.UserID != userID {
		return renderPage(c, "Ad Deleted", ui.AdDeleted())
	}

	// Update the ad category cookie based on the ad
	cookie.SetCategoryID(c, a.CategoryID)

	title := "Rocky Ads - " + a.Title
	csrfToken := local.GetCSRFToken(c)
	return renderPage(c, title, ui.Ad(adID, userID, a.ImageCount, a.Price,
		a.Title, a.Location(), a.Description, a.CreatedAt, a.Bookmarked,
		!a.IsDeleted(), csrfToken))
}

func NewAdHandler(c *fiber.Ctx) error {
	return renderPage(c, "New Ad", ui.NewAd())
}

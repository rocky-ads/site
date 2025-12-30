package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
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
	return renderPage(c, title, ui.Ad(adID, userID, a.UserID, a.ImageCount, a.Price,
		a.Title, a.Location(), a.Description, a.CreatedAt, a.Bookmarked,
		!a.IsDeleted(), csrfToken))
}

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

	fields, err := field.GetFields(categoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	fv := make(field.Values)
	renderedFields := make([]g.Node, 0, len(fields))
	for _, f := range fields {
		fieldData := f.GetField()
		// Skip fields that are in a chain but not first in chain
		// Only apply chain filtering to SpecFields; non-SpecFields (like price, location) should always be shown
		if fieldData.PrevFieldID != 0 {
			if specFielder, ok := f.(field.SpecFielder); ok {
				// Field is a SpecField in a chain, check if it's first
				specField := specFielder.GetSpecField()
				if !specField.IsFirst {
					// In chain but not first, skip
					continue
				}
			}
			// Not a SpecField - always show it regardless of PrevFieldID
		}
		renderedFields = append(renderedFields, f.NewAdNode(fv))
	}

	return renderPage(c, "New Ad", ui.NewAd(categoryName, renderedFields))
}

func AdShareHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	return render(c, ui.AdShareModal(fmt.Sprintf("/ad/%d/share", adID)))
}

func DeleteAdHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	userID := local.GetUserID(c)
	loc := cookie.GetLocation(c)

	// Get ad to verify ownership
	a, err := ad.GetAd(userID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	// Check ownership
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}

	// Delete the ad
	if err := deleteAd(adID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete ad")
	}

	// Redirect to the ad page
	c.Set("HX-Redirect", fmt.Sprintf("/ad/%d", adID))
	return c.SendStatus(fiber.StatusOK)
}

func RestoreAdHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	userID := local.GetUserID(c)
	loc := cookie.GetLocation(c)

	// Get ad to verify ownership
	a, err := ad.GetAd(userID, adID, loc)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	// Check ownership
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}

	// Restore the ad
	if err := restoreAd(adID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to restore ad")
	}

	// Redirect to the ad page
	c.Set("HX-Redirect", fmt.Sprintf("/ad/%d", adID))
	return c.SendStatus(fiber.StatusOK)
}

func deleteAd(adID int) error {
	_, err := db.Exec("UPDATE ads SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", adID)
	return err
}

func restoreAd(adID int) error {
	_, err := db.Exec("UPDATE ads SET deleted_at = NULL WHERE id = ?", adID)
	return err
}

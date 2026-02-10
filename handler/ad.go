package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/egg"
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

	title := config.ServerName + " - " + a.Title
	csrfToken := local.GetCSRFToken(c)
	// TODO: Determine reachable based on owner's contact info/verification status
	reachable := true // For now, assume owner is reachable

	return renderPage(c, title, ui.Ad(adID, userID, a.UserID, a.ImageCount, a.Price,
		a.Title, a.Location(), a.Description, a.CreatedAt, a.Bookmarked,
		!a.IsDeleted(), reachable, csrfToken, a.RockCount))
}

func AdEggConversationHandler(c *fiber.Ctx) error {
	adID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	ordinal, err := c.ParamsInt("ordinal")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid egg ordinal")
	}

	currentUserID := local.GetUserID(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	// Get conversation ID by ordinal
	conversationID, err := egg.GetPublicConversationIDByOrdinal(adID, ordinal)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Egg conversation not found")
	}

	// Get conversation
	conv, err := message.GetConversationByID(conversationID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	// Check if conversation is public or user is participant
	if conv.EggThrowerID == nil && conv.OwnerID != currentUserID && conv.EnquirerID != currentUserID {
		return fiber.NewError(fiber.StatusForbidden, "Conversation not found")
	}

	// Mark conversation as read when opened (only if user is participant)
	if conv.OwnerID == currentUserID || conv.EnquirerID == currentUserID {
		if err := message.MarkConversationAsRead(conversationID, currentUserID); err != nil {
			// Log error but don't fail the request
			logger.Error("Failed to mark conversation as read", "error", err, "conversationID", conversationID, "userID", currentUserID)
		}
	}

	return renderConversationModal(c, conv, currentUserID, loc, csrfToken)
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
		// Show only first field of each chain. PrevFieldID is per-category order (not per-chain), so we need IsFirst.
		if specFielder, ok := f.(field.SpecFielder); ok && !specFielder.GetSpecField().IsFirst {
			continue
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

	// Construct full URL
	protocol := "https"
	if c.Protocol() == "http" {
		protocol = "http"
	}
	path := fmt.Sprintf("%s://%s/ad/%d", protocol, c.Hostname(), adID)

	return render(c, ui.AdShareModal(path))
}

func AdShareCopyHandler(c *fiber.Ctx) error {
	path := c.Query("path", "")
	copied := c.Query("copied", "false") == "true"

	if copied {
		return render(c, ui.CopyButtonCopied(path))
	}
	return render(c, ui.CopyButton(path))
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

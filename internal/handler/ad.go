package handler

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/egg"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/param"
	"github.com/rocky-ads/site/internal/ui"
	"github.com/rocky-ads/site/internal/user"
	"github.com/rocky-ads/site/internal/vector"
)

func AdHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	userID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)

	a, err := ad.GetAd(userID, adID, tz)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	if err := ad.LoadTags(&a); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// If ad is deleted and user is not the owner, show deleted message
	if a.IsDeleted() && a.UserID != userID {
		return renderPage(c, "Ad Deleted", ui.AdDeleted())
	}

	// Update the ad category cookie based on the ad
	cookie.SetCategoryID(c, a.CategoryID)

	if !a.IsDeleted() && userID != 0 {
		_ = ad.IncrementAdClickForUser(adID, userID)
	}

	csrfToken := local.GetCSRFToken(c)

	reachable, err := user.IsReachable(a.UserID)
	if err != nil {
		reachable = false
	}

	isTest, err := user.IsTestUser(a.UserID)
	if err != nil {
		isTest = false
	}

	return renderPage(c, a.Title, ui.Ad(adDetailFrom(a, userID, reachable, isTest), userID, csrfToken))
}

func adDetailFrom(a ad.Ad, viewerUserID int, reachable, isTest bool) ui.AdDetail {
	price, priceCurrency, hasPrice := a.PriceValue()
	priceDisplay := ""
	if hasPrice {
		priceDisplay = currency.Format(price, priceCurrency)
	}
	desc := ad.ParseDescriptionForDisplay(a.Description)
	var history []ui.AdHistoryEntry
	if local.IsLoggedIn(viewerUserID) {
		history = make([]ui.AdHistoryEntry, len(desc.History))
		for i, e := range desc.History {
			history[i] = ui.AdHistoryEntry{
				Header:       e.Header,
				Body:         e.Body,
				ImageIndices: e.ImageIndices,
			}
		}
	}
	return ui.AdDetail{
		ID:                  a.ID,
		OwnerID:             a.UserID,
		ImageCount:          a.ImageCount,
		PriceDisplay:        priceDisplay,
		HasPrice:            hasPrice,
		Title:               a.Title,
		Location:            ad.AdLocationDisplay(a, viewerUserID),
		DescriptionOriginal: desc.Original,
		DescriptionHistory:  history,
		CreatedAt:           a.CreatedAt,
		Bookmarked:          a.Bookmarked,
		Active:              !a.IsDeleted(),
		IsTest:              isTest,
		Reachable:           reachable,
		RockCount:           a.RockCount,
		FacetLabel:          adFacetLabel(a),
		FacetDetails:        adDetailFacetDisplays(a),
		Tags:                adTagDisplays(a),
	}
}

func adTagDisplays(a ad.Ad) []string {
	if len(a.Tags) == 0 {
		return nil
	}
	out := make([]string, len(a.Tags))
	for i, s := range a.Tags {
		out[i] = s.Display()
	}
	return out
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
	tz := cookie.GetTimezone(c)
	csrfToken := local.GetCSRFToken(c)

	// Get conversation ID by ordinal
	conversationID, err := egg.GetPublicConversationIDByOrdinal(adID, ordinal)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Egg conversation not found")
	}

	conv, _, err := message.OpenConversation(conversationID, currentUserID)
	if errors.Is(err, message.ErrModalAccess) {
		return fiber.NewError(fiber.StatusForbidden, "Conversation not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Conversation not found")
	}

	if message.IsParticipant(conv, currentUserID) {
		return renderConversationModal(c, conv, currentUserID, tz, csrfToken)
	}
	return renderEggOpinionModal(c, conv, currentUserID, tz)
}

func AdShareHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	userID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)
	a, err := ad.GetAd(userID, adID, tz)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	if a.IsDeleted() && a.UserID != userID {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

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
	tz := cookie.GetTimezone(c)

	// Get ad to verify ownership
	a, err := ad.GetAd(userID, adID, tz)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	// Check ownership
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}

	// Delete the ad
	if err := ad.Delete(adID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete ad")
	}
	_ = vector.DeleteAdEmbedding(adID)

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
	tz := cookie.GetTimezone(c)

	// Get ad to verify ownership
	a, err := ad.GetAd(userID, adID, tz)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}

	// Check ownership
	if a.UserID != userID {
		return fiber.NewError(fiber.StatusForbidden, "You are not the owner of this ad")
	}

	// Restore the ad
	if err := ad.Restore(adID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to restore ad")
	}
	vector.QueueAd(adID)

	// Redirect to the ad page
	c.Set("HX-Redirect", fmt.Sprintf("/ad/%d", adID))
	return c.SendStatus(fiber.StatusOK)
}

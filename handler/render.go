package handler

import (
	"time"

	"maragu.dev/gomponents"
	g "maragu.dev/gomponents"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/search"
	"github.com/rocky-ads/site/ui"
)

// render renders a gomponents Node as HTML response
func render(c *fiber.Ctx, component g.Node) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return component.Render(c.Response().BodyWriter())
}

func renderPage(c *fiber.Ctx, title string, body []gomponents.Node) error {
	userID := local.GetUserID(c)
	userName := local.GetUserName(c)
	csrfToken := local.GetCSRFToken(c)
	return render(c, ui.Page(userID, userName, title, c.Path(), csrfToken, body))
}

// searchAndRenderAds searches for ads and renders them into gomponents nodes
func searchAndRenderAds(categoryID int, view int, fv field.Values, loc *time.Location) ([]g.Node, error) {
	adIDs, err := search.Search(categoryID, fv)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	results, err := ad.AdNodes(adIDs, view, loc)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return results, nil
}

package handler

import (
	"bytes"
	"time"

	"maragu.dev/gomponents"
	g "maragu.dev/gomponents"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/field"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/message"
	"github.com/rocky-ads/site/search"
	"github.com/rocky-ads/site/ui"
)

// render renders a gomponents Node as HTML response
func render(c *fiber.Ctx, component g.Node) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return component.Render(c.Response().BodyWriter())
}

// renderSVG renders a gomponents Node as SVG response
func renderSVG(c *fiber.Ctx, component g.Node) error {
	c.Set(fiber.HeaderContentType, "image/svg+xml")
	return component.Render(c.Response().BodyWriter())
}

func renderPage(c *fiber.Ctx, title string, body []gomponents.Node) error {
	userID := local.GetUserID(c)
	userName := local.GetUserName(c)
	csrfToken := local.GetCSRFToken(c)
	hasUnread, _ := message.GetHasUnread(userID)

	// Close existing SSE connections before rendering new page
	closeSSE(userID)

	return render(c, ui.Page(userID, hasUnread, userName, title, c.Path(), csrfToken, body))
}

// searchAndRenderAds searches for ads and renders them into gomponents nodes
func searchAndRenderAds(categoryID, limit, offset, userID, view int, fv field.Values, loc *time.Location, csrfToken string) ([]g.Node, error) {
	adIDs, err := search.Search(categoryID, limit, offset, fv)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	page := (offset / limit) + 1
	results, err := ad.AdNodes(adIDs, userID, view, page, loc, csrfToken, true)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return results, nil
}

// renderToString renders a gomponents Node to an HTML string
func renderToString(component g.Node) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

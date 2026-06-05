package handler

import (
	"bytes"
	"time"

	"maragu.dev/gomponents"
	g "maragu.dev/gomponents"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/search"
	"github.com/rocky-ads/site/internal/ui"
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

// searchFromRequest runs search for the current request page and renders ad nodes.
func searchFromRequest(c *fiber.Ctx, state cookie.SearchState) (view, page int, results []g.Node, err error) {
	categoryID := cookie.GetCategoryID(c)
	userID := local.GetUserID(c)
	view = cookie.GetView(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)
	page = c.QueryInt("page", 1)

	p := parseSearchParamsFromState(c, state, categoryID)
	results, err = searchAndRenderAds(p, userID, view, loc, csrfToken)
	return view, page, results, err
}

// searchAdsForUI runs search and returns presentation-ready ad cards.
func searchAdsForUI(p search.Params, userID int, loc *time.Location) ([]ui.AdCard, error) {
	adIDs, err := search.Search(p)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	ads, err := ad.GetAds(userID, adIDs, loc)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return adCardsFrom(ads, loc), nil
}

// searchAndRenderAds searches for ads and renders them into gomponents nodes.
func searchAndRenderAds(p search.Params, userID, view int, loc *time.Location, csrfToken string) ([]g.Node, error) {
	cards, err := searchAdsForUI(p, userID, loc)
	if err != nil {
		return nil, err
	}

	page := (p.Offset / p.Limit) + 1
	return ui.AdNodes(cards, userID, view, page, csrfToken, true), nil
}

func adCardFrom(a ad.Ad, loc *time.Location) ui.AdCard {
	price, priceCurrency, hasPrice := a.PriceValue()
	return ui.AdCardFromFields(
		a.ID, price, a.ImageCount, a.RockCount,
		priceCurrency, a.Title, a.Location(), a.FacetLabel(),
		hasPrice, a.CreatedAt.In(loc),
		!a.IsDeleted(), a.Bookmarked,
	)
}

func adCardsFrom(ads []ad.Ad, loc *time.Location) []ui.AdCard {
	cards := make([]ui.AdCard, len(ads))
	for i, a := range ads {
		cards[i] = adCardFrom(a, loc)
	}
	return cards
}

// renderToString renders a gomponents Node to an HTML string
func renderToString(component g.Node) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

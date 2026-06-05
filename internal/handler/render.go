package handler

import (
	"bytes"
	"strings"
	"time"

	"maragu.dev/gomponents"
	g "maragu.dev/gomponents"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/currency"
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
	view = ui.ValidateView(cookie.GetView(c))
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
	priceDisplay := ""
	if hasPrice {
		priceDisplay = currency.Format(price, priceCurrency)
	}
	return ui.AdCardFromFields(
		a.ID, a.ImageCount, a.RockCount,
		priceDisplay, a.Title, adLocation(a), adFacetLabel(a),
		hasPrice, a.CreatedAt.In(loc),
		!a.IsDeleted(), a.Bookmarked,
	)
}

// adLocation builds the display location string (flag + city/admin area) for an ad.
func adLocation(a ad.Ad) string {
	if a.City == "" && a.AdminArea == "" && a.Country == "" {
		return ""
	}

	var locationText string
	if a.City != "" && a.AdminArea != "" {
		locationText = a.City + ", " + a.AdminArea
	} else if a.City != "" {
		locationText = a.City
	} else if a.AdminArea != "" {
		locationText = a.AdminArea
	}

	var flag string
	if len(a.Country) == 2 {
		code := strings.ToUpper(a.Country)
		flag = string(rune(int32(code[0])-'A'+0x1F1E6)) + string(rune(int32(code[1])-'A'+0x1F1E6))
	}

	return flag + " " + locationText
}

// adFacetLabel returns the compact, non-price facet labels for a listing card,
// joined with separators (e.g. "45K mi · 2020").
func adFacetLabel(a ad.Ad) string {
	return strings.Join(adFacetLabels(a, true), " · ")
}

// adFacetLabels returns the non-price facet labels for an ad. When compact is
// true it uses each facet's compact format; otherwise the full format.
func adFacetLabels(a ad.Ad, compact bool) []string {
	cat, err := ad.GetCategory(a.CategoryID)
	if err != nil {
		return nil
	}

	var labels []string
	for _, d := range cat.Facets() {
		if d.Key == "price" {
			continue
		}
		v, ok := a.Facets[d.Key]
		if !ok {
			continue
		}
		var s string
		if compact {
			s = d.FormatCompact(v)
		} else {
			s = d.FormatFull(v)
		}
		if s != "" {
			labels = append(labels, s)
		}
	}
	return labels
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

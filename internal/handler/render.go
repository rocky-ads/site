package handler

import (
	"bytes"
	"strings"
	"time"

	"maragu.dev/gomponents"
	g "maragu.dev/gomponents"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/facet"
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

	showIntroBanner := !local.IsLoggedIn(userID) &&
		!cookie.IsIntroBannerDismissed(c) &&
		c.Path() != "/about"

	return render(c, ui.Page(userID, hasUnread, userName, title,
		c.Path(), csrfToken, showIntroBanner, body))
}

// searchFromRequest runs search for the current request page and renders ad nodes.
func searchFromRequest(c *fiber.Ctx,
	state cookie.SearchState) (view, page int, results []g.Node, err error) {
	view = ui.ValidateView(cookie.GetView(c))
	page, results, err = searchAndRender(c, state, cookie.GetCategoryID(c), view)
	return view, page, results, err
}

// searchAndRender runs search for the current page and returns rendered ad nodes.
func searchAndRender(c *fiber.Ctx, state cookie.SearchState, categoryID,
	view int) (page int, results []g.Node, err error) {

	userID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)
	csrfToken := local.GetCSRFToken(c)
	distanceUnit := cookie.GetDistanceUnit(c)
	page = c.QueryInt("page", 1)

	p := parseSearchParams(state, page, categoryID, userID, distanceUnit, tz)
	ads, inAreaCount, err := searchForAds(p, userID, tz)
	if err != nil {
		return 0, nil, err
	}

	location := searchLocationDisplay(state.Location)
	results = renderSearchResults(
		ads, inAreaCount, page, p.HasGeo,
		userID, view, tz, csrfToken,
		location, state.Within, distanceUnit,
	)

	return page, results, nil
}

// searchForAds runs vector search and loads matching ads in result order.
func searchForAds(p search.Params, userID int,
	tz *time.Location) ([]ad.Ad, int, error) {
	result, err := search.Search(p)
	if err != nil {
		return nil, 0,
			fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	ads, err := ad.GetAds(userID, result.IDs, tz)
	if err != nil {
		return nil, 0,
			fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return ads, result.InAreaCount, nil
}

// renderSearchResults renders search result ads into gomponents nodes.
func renderSearchResults(ads []ad.Ad, inAreaCount, page int, hasGeo bool,
	userID, view int, tz *time.Location, csrfToken,
	location string, within int, distanceUnit string) []g.Node {
	cards := adCardsFrom(ads, userID, tz)
	cardNodes := ui.AdNodes(cards, userID, view, page, csrfToken, true)

	if !hasGeo {
		return cardNodes
	}

	displayWithin := within
	if displayWithin == 0 {
		displayWithin = 25
	}

	offset := (page - 1) * config.SearchPageSize

	if offset == 0 && inAreaCount == 0 {
		nodes := []g.Node{ui.NoInAreaMatchesMessage(displayWithin, distanceUnit, location)}
		if len(cardNodes) > 0 {
			nodes = append(nodes, ui.OutsideAreaHeading())
			nodes = append(nodes, cardNodes...)
		}
		return nodes
	}

	headingAt := -1
	if offset < inAreaCount && offset+len(cards) > inAreaCount {
		headingAt = inAreaCount - offset
	} else if offset == inAreaCount && inAreaCount > 0 {
		headingAt = 0
	}

	if headingAt < 0 || headingAt >= len(cardNodes) {
		return cardNodes
	}

	nodes := make([]g.Node, 0, len(cardNodes)+1)
	nodes = append(nodes, cardNodes[:headingAt]...)
	nodes = append(nodes, ui.OutsideAreaHeading())
	nodes = append(nodes, cardNodes[headingAt:]...)
	return nodes
}

func adCardFrom(a ad.Ad, viewerUserID int, tz *time.Location) ui.AdCard {
	price, priceCurrency, hasPrice := a.PriceValue()
	priceDisplay := ""
	if hasPrice {
		priceDisplay = currency.Format(price, priceCurrency)
	}
	return ui.AdCardFromFields(
		a.ID, a.ImageCount, a.RockCount,
		priceDisplay, a.Title,
		ad.AdLocationDisplay(a, viewerUserID), adFacetLabel(a),
		hasPrice, a.CreatedAt.In(tz),
		!a.IsDeleted(), a.Bookmarked,
	)
}

// adFacetLabel returns the compact, non-price facet labels for a listing card,
// joined with separators (e.g. "45K mi · 2020").
func adFacetLabel(a ad.Ad) string {
	return strings.Join(adFacetLabels(a, true), " · ")
}

// adDetailFacetDisplays returns formal facet pills for the ad detail page,
// excluding facets already shown in the title (mileage/hours) or price row.
func adDetailFacetDisplays(a ad.Ad) []string {
	cat := ad.GetCategory(a.CategoryID)
	var labels []string
	for _, d := range cat.Facets() {
		if d.Key == "price" || d.Kind == facet.Location {
			continue
		}
		if d.Key == "sale_end_date" {
			continue
		}
		if d.Key == "sale_start_date" {
			start := ""
			if v, ok := a.Facets["sale_start_date"]; ok {
				start = v.DateString()
			}
			end := ""
			if v, ok := a.Facets["sale_end_date"]; ok {
				end = v.DateString()
			}
			labels = append(labels, facet.SaleDateDetailDisplays(start, end)...)
			continue
		}
		if d.CardLabel() {
			continue
		}
		v, ok := a.Facets[d.Key]
		if !ok || !v.Present() {
			continue
		}
		s := d.FormatFull(v)
		if s == "" {
			continue
		}
		labels = append(labels, d.Label+": "+s)
	}
	return labels
}

// adFacetLabels returns the non-price facet labels for an ad. When compact is
// true it keeps only card facets (e.g. mileage) and uses compact formatting.
func adFacetLabels(a ad.Ad, compact bool) []string {
	cat := ad.GetCategory(a.CategoryID)

	var labels []string
	for _, d := range cat.Facets() {
		if d.Key == "price" {
			continue
		}
		if compact {
			if d.Key == "sale_end_date" {
				continue
			}
			if !d.CardLabel() {
				continue
			}
			if d.Key == "sale_start_date" {
				start := ""
				if v, ok := a.Facets["sale_start_date"]; ok {
					start = v.DateString()
				}
				end := ""
				if v, ok := a.Facets["sale_end_date"]; ok {
					end = v.DateString()
				}
				if s := facet.FormatDateRange(start, end); s != "" {
					labels = append(labels, s)
				}
				continue
			}
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

func adCardsFrom(ads []ad.Ad, viewerUserID int, tz *time.Location) []ui.AdCard {
	cards := make([]ui.AdCard, len(ads))
	for i, a := range ads {
		cards[i] = adCardFrom(a, viewerUserID, tz)
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

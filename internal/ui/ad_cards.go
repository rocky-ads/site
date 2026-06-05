package ui

import (
	"time"

	g "maragu.dev/gomponents"
)

// AdCardFromFields builds an AdCard from presentation field values.
func AdCardFromFields(
	id, price, imageCount, rockCount int,
	priceCurrency, title, location, facetLabel string,
	createdAt time.Time,
	active, bookmarked bool,
) AdCard {
	return AdCard{
		ID:            id,
		Price:         price,
		PriceCurrency: priceCurrency,
		Title:         title,
		Location:      location,
		FacetLabel:    facetLabel,
		CreatedAt:     createdAt,
		ImageCount:    imageCount,
		Active:        active,
		Bookmarked:    bookmarked,
		RockCount:     rockCount,
	}
}

// AdNodes renders search or list ad cards into gomponents nodes.
func AdNodes(cards []AdCard, userID, view, page int, csrfToken string, pagination bool) []g.Node {
	results := make([]g.Node, len(cards))
	for i, card := range cards {
		isLast := i == len(cards)-1 && pagination
		results[i] = adCardNode(card, userID, view, page+1, csrfToken, isLast)
	}
	return results
}

func adCardNode(card AdCard, userID, view, nextPage int, csrfToken string, isLast bool) g.Node {
	switch view {
	case ViewGrid:
		return AdGridNode(
			userID, card.ID, card.Price, card.ImageCount, nextPage,
			card.PriceCurrency, card.Title, card.Location, card.FacetLabel, csrfToken,
			card.CreatedAt, card.Active, card.Bookmarked, isLast, card.RockCount,
		)
	case ViewList:
		return AdListNode(
			userID, card.ID, card.Price, card.PriceCurrency, card.Title, card.Location, card.FacetLabel,
			card.CreatedAt, card.Active, card.Bookmarked, csrfToken, isLast, nextPage,
			card.RockCount,
		)
	default:
		return g.Text("bad view")
	}
}

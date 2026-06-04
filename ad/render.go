package ad

import (
	"time"

	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

func AdNodes(adIDs []int, userID, view, page int, loc *time.Location, csrfToken string, pagination bool) ([]g.Node, error) {
	ads, err := GetAds(userID, adIDs, loc)
	if err != nil {
		return nil, err
	}

	results := make([]g.Node, len(ads))
	for i, ad := range ads {
		isLast := i == len(ads)-1 && pagination
		results[i] = ad.Node(userID, view, page+1, csrfToken, isLast)
	}
	return results, nil
}

func (a Ad) Node(userID, view, nextPage int, csrfToken string, isLast bool) g.Node {
	switch view {
	case ui.ViewGrid:
		return ui.AdGridNode(userID, a.ID, a.Price, a.ImageCount, nextPage, a.PriceCurrency, a.Title, a.Location(), csrfToken, a.CreatedAt, !a.IsDeleted(), a.Bookmarked, isLast, a.RockCount)
	case ui.ViewList:
		return ui.AdListNode(userID, a.ID, a.Price, a.PriceCurrency, a.Title, a.Location(), a.CreatedAt, !a.IsDeleted(), a.Bookmarked, csrfToken, isLast, nextPage, a.RockCount)
	default:
		return g.Text("bad view")
	}
}

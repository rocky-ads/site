package ad

import (
	"time"

	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

func AdNodes(adIDs []int, userID int, view int, loc *time.Location, csrfToken string) ([]g.Node, error) {
	ads, err := GetAds(userID, adIDs, loc)
	if err != nil {
		return nil, err
	}
	results := make([]g.Node, len(ads))
	for i, ad := range ads {
		results[i] = ad.Node(userID, view, csrfToken)
	}
	return results, nil
}

func (a Ad) Node(userID int, view int, csrfToken string) g.Node {
	switch view {
	case ui.ViewGrid:
		return ui.AdGridNode(a.ID, a.Title)
	case ui.ViewList:
		return ui.AdListNode(userID, a.ID, a.Title, !a.IsDeleted(), a.Bookmarked, csrfToken)
	case ui.ViewTree:
		return ui.AdTreeNode(a.ID, a.Title)
	default:
		return g.Text("bad view")
	}
}

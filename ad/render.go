package ad

import (
	"time"

	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

func AdNodes(adIDs []int, view int, loc *time.Location) ([]g.Node, error) {
	ads, err := GetAds(adIDs, loc)
	if err != nil {
		return nil, err
	}
	results := make([]g.Node, len(ads))
	for i, ad := range ads {
		results[i] = ad.Node(view)
	}
	return results, nil
}

func (a Ad) Node(view int) g.Node {
	switch view {
	case ui.ViewGrid:
		return ui.AdGridNode(a.ID, a.Title)
	case ui.ViewList:
		return ui.AdListNode(a.ID, a.Title)
	case ui.ViewTree:
		return ui.AdTreeNode(a.ID, a.Title)
	default:
		return g.Text("bad view")
	}
}

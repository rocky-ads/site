package ad

import (
	"strings"
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

func (a Ad) location() string {
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

func (a Ad) Node(userID int, view int, csrfToken string) g.Node {
	switch view {
	case ui.ViewGrid:
		return ui.AdGridNode(userID, a.ID, a.Price, a.ImageCount, a.Title, a.location(), a.CreatedAt, !a.IsDeleted(), a.Bookmarked, csrfToken)
	case ui.ViewList:
		return ui.AdListNode(userID, a.ID, a.Price, a.Title, a.location(), a.CreatedAt, !a.IsDeleted(), a.Bookmarked, csrfToken)
	case ui.ViewTree:
		return ui.AdTreeNode(a.ID, a.Title)
	default:
		return g.Text("bad view")
	}
}

package ad

import (
	"strings"
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
	var nextPage int
	if pagination {
		nextPage = page + 1
	}
	for i, ad := range ads {
		isLast := i == len(ads)-1 && pagination
		results[i] = ad.Node(userID, view, nextPage, csrfToken, isLast)
	}
	return results, nil
}

func (a Ad) Location() string {
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

func (a Ad) Node(userID, view, nextPage int, csrfToken string, isLast bool) g.Node {
	switch view {
	case ui.ViewGrid:
		return ui.AdGridNode(userID, a.ID, a.Price, a.ImageCount, nextPage, a.Title, a.Location(), csrfToken, a.CreatedAt, !a.IsDeleted(), a.Bookmarked, isLast)
	case ui.ViewList:
		return ui.AdListNode(userID, a.ID, a.Price, a.Title, a.Location(), a.CreatedAt, !a.IsDeleted(), a.Bookmarked, csrfToken, isLast, nextPage)
	default:
		return g.Text("bad view")
	}
}

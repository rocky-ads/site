package ad

import (
	"fmt"

	g "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func RenderAds(adIDs []int) ([]g.Node, error) {
	ads, err := GetAds(adIDs)
	if err != nil {
		return nil, err
	}
	results := make([]g.Node, len(ads))
	for i, ad := range ads {
		results[i] = ad.Render()
	}
	return results, nil
}

func (a Ad) Render() g.Node {
	return Div(
		ID(fmt.Sprintf("ad-%d", a.ID)),
		g.Text(a.Title),
	)
}

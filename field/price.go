package field

import (
	g "maragu.dev/gomponents"

	"github.com/rocky-ads/site/ui"
)

type PriceField struct {
	Field
}

func (f PriceField) FilterNode(fv Values) g.Node {
	minPrice := fv.Get("min_price")
	maxPrice := fv.Get("max_price")

	return ui.PriceRange(f.DisplayName, minPrice, maxPrice)
}

func (f PriceField) NewAdNode(fv Values, opts NewAdOpts) g.Node {
	return ui.PriceInput(f.IsRequired, opts.DefaultCurrency)
}

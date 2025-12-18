package field

import (
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type LocationField struct {
	Field
}

func (f LocationField) FilterNode(fv Values) g.Node {

	location := fv.Get("location")
	radius := fv.Get("radius")

	if radius == "" {
		radius = "25"
	}

	return ui.LocationRadius(location, radius)
}

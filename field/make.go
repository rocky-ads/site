package field

import (
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type FieldMake struct {
	SpecField
}

func (f FieldMake) FilterNode(fv Values) g.Node {

	values, err := f.getAnyValues(fv)
	if err != nil {
		logger.Error("Failed to get values for make field: %w", err)
		return nil
	}

	value := fv.Get("make")

	return ui.FieldSelect(f.Name, f.DisplayName, value, values)
}

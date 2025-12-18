package field

import (
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type PartCategoryField struct {
	SpecField
}

func (f PartCategoryField) FilterNode(fv Values) g.Node {

	values, err := f.GetAnyValues(fv)
	if err != nil {
		logger.Error("Failed to get values for make field: %w", err)
		return nil
	}

	value := fv.Get("part_category")

	return ui.FieldSelect(f.Name, f.DisplayName, value, values)
}

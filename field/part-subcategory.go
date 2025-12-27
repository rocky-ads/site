package field

import (
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type PartSubcategoryField struct {
	SpecField
}

func (f PartSubcategoryField) FilterNode(fv Values) g.Node {
	values, err := f.GetAnyValues(fv)
	if err != nil {
		logger.Error("Failed to get values for part_subcategory field: %w", err)
		return nil
	}

	value := fv.Get("part_subcategory")
	return ui.FieldSelect(f.Name, f.DisplayName, value, values)
}

func (f PartSubcategoryField) NewAdNode(fv Values) g.Node {
	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for part_subcategory field: %w", err)
		return nil
	}

	return ui.FieldSelect(f.Name, f.DisplayName, "", values)
}

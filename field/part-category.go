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
		logger.Error("Failed to get values for field", "field", f.Name, "error", err)
		return nil
	}

	value := fv.Get("part_category")

	return ui.FieldSelect(f.Name, f.DisplayName, value, "", "", values, false)
}

func (f PartCategoryField) NewAdNode(fv Values, opts NewAdOpts) g.Node {
	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for field", "field", f.Name, "error", err)
		return nil
	}

	return ui.FieldSelect(f.Name, f.DisplayName, "", f.NextFieldName, fv.Encode(), values, f.IsRequired)
}

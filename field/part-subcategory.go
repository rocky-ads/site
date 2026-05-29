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
	return nil
}

func (f PartSubcategoryField) NewAdNode(fv Values, opts NewAdOpts) g.Node {
	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for field", "field", f.Name, "error", err)
		return nil
	}

	return ui.FieldSelect(f.Name, f.DisplayName, "", f.NextFieldName, fv.Encode(), values, f.IsRequired)
}

package field

import (
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type ModelField struct {
	SpecField
}

func (f ModelField) FilterNode(fv Values) g.Node {
	return nil
}

func (f ModelField) NewAdNode(fv Values, opts NewAdOpts) g.Node {
	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for field", "field", f.Name, "error", err)
		return nil
	}

	return ui.FieldCheckboxes(f.Name, f.DisplayName, f.NextFieldName, fv.Encode(), values)
}

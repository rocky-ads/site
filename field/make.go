package field

import (
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type MakeField struct {
	SpecField
}

func (f MakeField) FilterNode(fv Values) g.Node {

	values, err := f.GetAnyValues(fv)
	if err != nil {
		logger.Error("Failed to get values for field", "field", f.Name, "error", err)
		return nil
	}

	value := fv.Get("make")

	return ui.FieldSelect(f.Name, f.DisplayName, value, "", "", values, false)
}

func (f MakeField) NewAdNode(fv Values) g.Node {

	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for field", "field", f.Name, "error", err)
		return nil
	}

	return ui.FieldSelect(f.Name, f.DisplayName, "", f.NextFieldName, fv.Encode(), values, f.IsRequired)
}

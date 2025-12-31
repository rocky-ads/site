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

func (f ModelField) NewAdNode(fv Values) g.Node {
	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for model field: %w", err)
		return nil
	}

	return ui.FieldSelect(f.Name, f.DisplayName, "", values)
}

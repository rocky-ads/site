package field

import (
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
	g "maragu.dev/gomponents"
)

type EngineField struct {
	SpecField
}

func (f EngineField) FilterNode(fv Values) g.Node {
	values, err := f.GetAnyValues(fv)
	if err != nil {
		logger.Error("Failed to get values for engine field: %w", err)
		return nil
	}

	value := fv.Get("engine")
	return ui.FieldSelect(f.Name, f.DisplayName, value, values)
}

func (f EngineField) NewAdNode(fv Values) g.Node {
	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for engine field: %w", err)
		return nil
	}

	return ui.FieldSelect(f.Name, f.DisplayName, "", values)
}

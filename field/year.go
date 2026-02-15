package field

import (
	"fmt"
	"time"

	g "maragu.dev/gomponents"

	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/ui"
)

type YearField struct {
	SpecField
}

func (f YearField) FilterNode(fv Values) g.Node {
	minYear := fv.Get("min_year")
	maxYear := fv.Get("max_year")

	// Calculate max year as current year + 2
	maxYearValue := time.Now().Year() + 2
	maxMaxYear := fmt.Sprintf("%d", maxYearValue)

	return ui.YearRange(f.DisplayName, minYear, maxYear, maxMaxYear)
}

func (f YearField) NewAdNode(fv Values) g.Node {
	values, err := f.GetAllValues(fv)
	if err != nil {
		logger.Error("Failed to get values for field", "field", f.Name, "error", err)
		return nil
	}

	return ui.FieldCheckboxes(f.Name, f.DisplayName, f.NextFieldName, fv.Encode(), values)
}

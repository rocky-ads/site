package field

var FilterExcludedFieldNames = map[string]struct{}{
	"title":       {},
	"description": {},
}

func IsFilterExcluded(fieldName string) bool {
	_, ok := FilterExcludedFieldNames[fieldName]
	return ok
}

type AdFilters struct {
	PriceMin *int
	PriceMax *int
	Location string
	Fields   map[string][]string
}

func (f AdFilters) Values(fieldName string) []string {
	if f.Fields == nil {
		return nil
	}
	return f.Fields[fieldName]
}

func (f AdFilters) HasFilters() bool {
	if f.PriceMin != nil || f.PriceMax != nil || f.Location != "" {
		return true
	}
	for _, vals := range f.Fields {
		if len(vals) > 0 {
			return true
		}
	}
	return false
}

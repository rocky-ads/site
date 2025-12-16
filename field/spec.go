package field

// SpecField represents a specification field (a field in a spec table)
type SpecField struct {
	Field
	SpecTable string `json:"spec_table"`
}

func (f SpecField) getAnyValues(fv Values) ([]string, error) {
	return f.Field.FilterNode(fv)
}

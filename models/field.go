package models

import "net/url"

type Field struct {
	ID          int    `json:"id"`
	Name        string `json:"field_name"`
	DisplayName string `json:"display_name"`
	IsRequired  bool   `json:"is_required"`
}

// SpecField represents a specification field (a field in a spec table)
type SpecField struct {
	Field
	SpecTable string `json:"spec_table"`
}

// FieldValues represents field values as URL values
type FieldValues = url.Values // map[string][]string

// FieldValuesByIDs represents field values keyed by field ID instead of field name
type FieldValuesByIDs map[int][]string

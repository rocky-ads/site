package models

import "net/url"

// Field represents a field in the system
type Field struct {
	ID          int    `json:"id"`
	Name        string `json:"field_name"`
	DisplayName string `json:"display_name,omitempty"`
	IsRequired  bool   `json:"is_required,omitempty"`
}

// SpecField represents a specification field (a field with fixed values from spec*.json files)
type SpecField struct {
	Field
	SpecTable string `json:"spec_table,omitempty"` // spec_table for the chain (empty string if not a spec chain)
}

// FieldValues represents field values as URL values
type FieldValues url.Values // map[string][]string

// FieldValuesByIDs represents field values keyed by field ID instead of field name
type FieldValuesByIDs map[int][]string

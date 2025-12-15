package services

import (
	"fmt"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/models"
)

var (
	fieldIDCache   = make(map[string]int)
	fieldCache     = make(map[string]models.Field)
	specFieldCache = make(map[categoryFieldKey]models.SpecField)
)

func GetFieldByName(fieldName string) (models.Field, error) {
	if field, ok := fieldCache[fieldName]; ok {
		return field, nil
	}

	var fields []models.Field
	query := `SELECT COALESCE(json_group_array(json_object('id', id,
		'field_name', name, 'display_name', display_name)), '[]') FROM fields ORDER BY id`
	if err := db.QueryJSON(&fields, query); err != nil {
		return models.Field{}, fmt.Errorf("loading fields: %w", err)
	}

	for _, f := range fields {
		fieldCache[f.Name] = f
	}

	if field, ok := fieldCache[fieldName]; ok {
		return field, nil
	}
	return models.Field{}, fmt.Errorf("field not found: %s", fieldName)
}

func GetFieldIDByName(fieldName string) (int, error) {
	if fieldID, ok := fieldIDCache[fieldName]; ok {
		return fieldID, nil
	}

	type fieldIDData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var fields []fieldIDData
	query := `SELECT COALESCE(json_group_array(json_object('id', id, 'name', name)), '[]') FROM fields ORDER BY id`
	if err := db.QueryJSON(&fields, query); err != nil {
		return 0, fmt.Errorf("loading fields: %w", err)
	}

	for _, f := range fields {
		fieldIDCache[f.Name] = f.ID
	}

	if fieldID, ok := fieldIDCache[fieldName]; ok {
		return fieldID, nil
	}
	return 0, fmt.Errorf("field not found: %s", fieldName)
}

func GetSpecField(categoryID int, fieldName string) (models.SpecField, error) {
	key := categoryFieldKey{categoryID: categoryID, fieldName: fieldName}
	if specField, ok := specFieldCache[key]; ok {
		return specField, nil
	}

	field, err := GetFieldByName(fieldName)
	if err != nil {
		return models.SpecField{}, err
	}

	specTable, err := GetSpecTable(categoryID, fieldName)
	if err != nil {
		return models.SpecField{}, err
	}

	specField := models.SpecField{
		Field:     field,
		SpecTable: specTable,
	}

	specFieldCache[key] = specField
	return specField, nil
}

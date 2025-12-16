package services

import (
	"fmt"
	"sort"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/models"
)

var (
	fieldCache           = make(map[string]models.Field)
	specFieldCache       = make(map[categoryFieldKey]models.SpecField)
	firstSpecFieldsCache = make(map[int][]models.SpecField)
	lastSpecFieldCache   = make(map[int]models.SpecField)
)

func GetFields(categoryID int) ([]models.Field, error) {
	var fields []models.Field
	query := `SELECT COALESCE(json_group_array(json_object('id', id,
		'field_name', name, 'display_name', display_name)), '[]') FROM fields ORDER BY id`
	if err := db.QueryJSON(&fields, query); err != nil {
		return nil, fmt.Errorf("loading fields: %w", err)
	}
	return fields, nil
}

func GetField(fieldName string) (models.Field, error) {
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

func GetSpecField(categoryID int, fieldName string) (models.SpecField, error) {
	key := categoryFieldKey{categoryID: categoryID, fieldName: fieldName}
	if specField, ok := specFieldCache[key]; ok {
		return specField, nil
	}

	field, err := GetField(fieldName)
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

func FirstSpecFields(categoryID int) ([]models.SpecField, error) {
	if cached, ok := firstSpecFieldsCache[categoryID]; ok {
		return cached, nil
	}

	type fieldData struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		ChainIndex  int    `json:"chain_index"`
		SpecTable   string `json:"spec_table"`
	}
	var allFields []fieldData
	query := `
		SELECT COALESCE(json_group_array(json_object(
			'id', f.id,
			'name', f.name,
			'display_name', f.display_name,
			'chain_index', c.chain_index,
			'spec_table', COALESCE(c.spec_table, '')
		)), '[]')
		FROM chain_fields cf
		JOIN chains c ON cf.chain_id = c.id
		JOIN fields f ON cf.field_id = f.id
		WHERE c.category_id = ? AND c.spec_table IS NOT NULL AND c.spec_table != '' AND cf.field_order = 0
		ORDER BY c.chain_index, cf.field_order
	`
	if err := db.QueryJSON(&allFields, query, categoryID); err != nil {
		return nil, err
	}

	firstFields := make(map[int]models.SpecField)
	seenChains := make(map[int]bool)

	for _, fd := range allFields {
		if !seenChains[fd.ChainIndex] {
			firstFields[fd.ChainIndex] = models.SpecField{
				Field: models.Field{
					ID:          fd.ID,
					Name:        fd.Name,
					DisplayName: fd.DisplayName,
				},
				SpecTable: fd.SpecTable,
			}
			seenChains[fd.ChainIndex] = true
		}
	}

	var result []models.SpecField
	var chainIndices []int
	for ci := range firstFields {
		chainIndices = append(chainIndices, ci)
	}
	sort.Ints(chainIndices)
	for _, ci := range chainIndices {
		result = append(result, firstFields[ci])
	}

	firstSpecFieldsCache[categoryID] = result
	return result, nil
}

func LastSpecField(categoryID int) (models.SpecField, error) {
	if cached, ok := lastSpecFieldCache[categoryID]; ok {
		return cached, nil
	}

	type allFieldData struct {
		ChainIndex int `json:"chain_index"`
		FieldOrder int `json:"field_order"`
	}
	var allChainFields []allFieldData
	queryAll := `
		SELECT COALESCE(json_group_array(json_object(
			'chain_index', c.chain_index,
			'field_order', cf.field_order
		)), '[]')
		FROM chain_fields cf
		JOIN chains c ON cf.chain_id = c.id
		WHERE c.category_id = ? AND c.spec_table IS NOT NULL AND c.spec_table != ''
		ORDER BY c.chain_index, cf.field_order
	`
	if err := db.QueryJSON(&allChainFields, queryAll, categoryID); err != nil {
		return models.SpecField{}, fmt.Errorf("finding last spec field: %w", err)
	}

	if len(allChainFields) == 0 {
		return models.SpecField{}, fmt.Errorf("finding last spec field: no field found")
	}

	chainFieldCounts := make(map[int]int)
	for _, fd := range allChainFields {
		chainFieldCounts[fd.ChainIndex]++
	}

	var chainIndices []int
	for ci := range chainFieldCounts {
		chainIndices = append(chainIndices, ci)
	}
	sort.Ints(chainIndices)

	chainOffsets := make(map[int]int)
	offset := 0
	for _, actualIdx := range chainIndices {
		chainOffsets[actualIdx] = offset
		offset += chainFieldCounts[actualIdx]
	}

	type lastFieldData struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		SpecTable   string `json:"spec_table"`
		ChainIndex  int    `json:"chain_index"`
		FieldOrder  int    `json:"field_order"`
	}
	var lastFields []lastFieldData
	queryLast := `
		SELECT COALESCE(json_group_array(json_object(
			'id', f.id,
			'name', f.name,
			'display_name', f.display_name,
			'spec_table', COALESCE(c.spec_table, ''),
			'chain_index', c.chain_index,
			'field_order', cf.field_order
		)), '[]')
		FROM chain_fields cf
		JOIN chains c ON cf.chain_id = c.id
		JOIN fields f ON cf.field_id = f.id
		WHERE c.category_id = ? AND c.spec_table IS NOT NULL AND c.spec_table != '' AND cf.next_in_chain = 0
		ORDER BY c.chain_index, cf.field_order
	`
	if err := db.QueryJSON(&lastFields, queryLast, categoryID); err != nil {
		return models.SpecField{}, fmt.Errorf("finding last spec field: %w", err)
	}

	if len(lastFields) == 0 {
		return models.SpecField{}, fmt.Errorf("finding last spec field: no field found")
	}

	var lastField lastFieldData
	maxAbsoluteOrder := 0
	for _, fd := range lastFields {
		absoluteOrder := chainOffsets[fd.ChainIndex] + fd.FieldOrder + 1
		if absoluteOrder > maxAbsoluteOrder {
			maxAbsoluteOrder = absoluteOrder
			lastField = fd
		}
	}

	result := models.SpecField{
		Field: models.Field{
			ID:          lastField.ID,
			Name:        lastField.Name,
			DisplayName: lastField.DisplayName,
		},
		SpecTable: lastField.SpecTable,
	}

	lastSpecFieldCache[categoryID] = result
	return result, nil
}

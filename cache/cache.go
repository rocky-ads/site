package cache

import (
	"fmt"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/models"
)

// CategoryFieldInfo holds cached category field metadata
type CategoryFieldInfo struct {
	NextInChain int
	ChainID     int // ID of the chain this field belongs to
	ChainIndex  int // Index of chain within category (for backward compatibility)
	FieldOrder  int
	SpecTable   string
}

type CategoryFieldKey struct {
	CategoryID int
	FieldName  string
}

type NextFieldKey struct {
	ChainID    int
	FieldOrder int
}

var (
	CategoryFieldCache       = make(map[CategoryFieldKey]CategoryFieldInfo)
	FieldIDCache             = make(map[string]int)
	FieldDisplayCache        = make(map[string]string)
	NextFieldCache           = make(map[NextFieldKey]string)
	VehicleTablePatternCache = make(map[int]string)
	CategoryCache            = make(map[string]int)
	FieldCache               = make(map[string]models.Field)
)

func Init() error {
	if err := initCategoryCache(); err != nil {
		return fmt.Errorf("initializing category cache: %w", err)
	}

	if err := initFieldCache(); err != nil {
		return fmt.Errorf("initializing field cache: %w", err)
	}

	if err := initFieldIDCache(); err != nil {
		return fmt.Errorf("initializing field ID cache: %w", err)
	}

	if err := initCategoryFieldCache(); err != nil {
		return fmt.Errorf("initializing category field cache: %w", err)
	}

	if err := initNextFieldCache(); err != nil {
		return fmt.Errorf("initializing next field cache: %w", err)
	}
	if err := initVehicleTablePatternCache(); err != nil {
		return fmt.Errorf("initializing vehicle table pattern cache: %w", err)
	}

	return nil
}

func initCategoryCache() error {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return fmt.Errorf("loading categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return fmt.Errorf("scanning category: %w", err)
		}
		CategoryCache[name] = id
	}

	return nil
}

func initFieldCache() error {
	rows, err := db.Query("SELECT id, name, display_name FROM fields")
	if err != nil {
		return fmt.Errorf("loading fields: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, displayName string
		if err := rows.Scan(&id, &name, &displayName); err != nil {
			return fmt.Errorf("scanning field: %w", err)
		}
		FieldCache[name] = models.Field{
			ID:          id,
			Name:        name,
			DisplayName: displayName,
		}
	}

	return nil
}

func initFieldIDCache() error {
	rows, err := db.Query("SELECT id, name, display_name FROM fields")
	if err != nil {
		return fmt.Errorf("loading fields: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, displayName string
		if err := rows.Scan(&id, &name, &displayName); err != nil {
			return fmt.Errorf("scanning field: %w", err)
		}
		FieldIDCache[name] = id
		FieldDisplayCache[name] = displayName
	}

	return nil
}

func initCategoryFieldCache() error {
	rows, err := db.Query(`
		SELECT c.category_id, f.name, cf.next_in_chain, c.id, c.chain_index, cf.field_order, COALESCE(c.spec_table, '') as spec_table
		FROM chain_fields cf
		JOIN chains c ON cf.chain_id = c.id
		JOIN fields f ON cf.field_id = f.id
	`)
	if err != nil {
		return fmt.Errorf("loading chain fields: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var categoryID int
		var fieldName string
		var info CategoryFieldInfo
		if err := rows.Scan(&categoryID, &fieldName, &info.NextInChain, &info.ChainID, &info.ChainIndex, &info.FieldOrder, &info.SpecTable); err != nil {
			return fmt.Errorf("scanning chain field: %w", err)
		}
		key := CategoryFieldKey{CategoryID: categoryID, FieldName: fieldName}
		CategoryFieldCache[key] = info
	}

	return nil
}

func initNextFieldCache() error {
	rows, err := db.Query(`
		SELECT cf.chain_id, cf.field_order, f.name
		FROM chain_fields cf
		JOIN fields f ON cf.field_id = f.id
	`)
	if err != nil {
		return fmt.Errorf("loading field orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chainID, fieldOrder int
		var fieldName string
		if err := rows.Scan(&chainID, &fieldOrder, &fieldName); err != nil {
			return fmt.Errorf("scanning field order: %w", err)
		}
		key := NextFieldKey{ChainID: chainID, FieldOrder: fieldOrder}
		NextFieldCache[key] = fieldName
	}

	return nil
}

func initVehicleTablePatternCache() error {
	rows, err := db.Query(`
		SELECT c.category_id, c.spec_table
		FROM chains c
		WHERE c.spec_table IS NOT NULL AND c.spec_table != ''
	`)
	if err != nil {
		return fmt.Errorf("loading spec table patterns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var categoryID int
		var pattern string
		if err := rows.Scan(&categoryID, &pattern); err != nil {
			return fmt.Errorf("scanning spec table pattern: %w", err)
		}
		if _, exists := VehicleTablePatternCache[categoryID]; !exists {
			VehicleTablePatternCache[categoryID] = pattern
		}
	}

	return nil
}

func GetCategoryIDByName(categoryName string) (int, error) {
	categoryID, ok := CategoryCache[categoryName]
	if !ok {
		return 0, fmt.Errorf("category not found: %s", categoryName)
	}
	return categoryID, nil
}

func GetFieldByName(fieldName string) (models.Field, error) {
	field, ok := FieldCache[fieldName]
	if !ok {
		return models.Field{}, fmt.Errorf("field not found: %s", fieldName)
	}
	return field, nil
}

func GetSpecTable(categoryID int, fieldName string) (string, error) {
	key := CategoryFieldKey{CategoryID: categoryID, FieldName: fieldName}
	if info, ok := CategoryFieldCache[key]; ok {
		return info.SpecTable, nil
	}
	return "", fmt.Errorf("field %s not found for category %d", fieldName, categoryID)
}

func GetSpecField(categoryID int, fieldName string) (models.SpecField, error) {
	field, err := GetFieldByName(fieldName)
	if err != nil {
		return models.SpecField{}, err
	}

	specTable, err := GetSpecTable(categoryID, fieldName)
	if err != nil {
		return models.SpecField{}, err
	}

	return models.SpecField{
		Field:     field,
		SpecTable: specTable,
	}, nil
}

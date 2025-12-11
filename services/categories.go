package services

import (
	"fmt"

	"github.com/rocky-ads/site/db"
)

type categoryFieldKey struct {
	categoryID int
	fieldName  string
}

var (
	categoryCache  = make(map[string]int)
	specTableCache = make(map[categoryFieldKey]string)
)

func GetCategoryIDByName(categoryName string) (int, error) {
	if categoryID, ok := categoryCache[categoryName]; ok {
		return categoryID, nil
	}

	type categoryData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var categories []categoryData
	query := `SELECT COALESCE(json_group_array(json_object('id', id, 'name', name)), '[]') FROM categories`
	if err := db.QueryJSON(&categories, query); err != nil {
		return 0, fmt.Errorf("loading categories: %w", err)
	}

	for _, cat := range categories {
		categoryCache[cat.Name] = cat.ID
	}

	if categoryID, ok := categoryCache[categoryName]; ok {
		return categoryID, nil
	}
	return 0, fmt.Errorf("category not found: %s", categoryName)
}

func GetSpecTable(categoryID int, fieldName string) (string, error) {
	key := categoryFieldKey{categoryID: categoryID, fieldName: fieldName}
	if specTable, ok := specTableCache[key]; ok {
		return specTable, nil
	}

	var specTables []string
	query := `
		SELECT COALESCE(json_group_array(COALESCE(c.spec_table, '')), '[]')
		FROM chain_fields cf
		JOIN chains c ON cf.chain_id = c.id
		JOIN fields f ON cf.field_id = f.id
		WHERE c.category_id = ? AND f.name = ?
		LIMIT 1
	`
	if err := db.QueryJSON(&specTables, query, categoryID, fieldName); err != nil {
		return "", fmt.Errorf("loading spec table: %w", err)
	}

	if len(specTables) == 0 || specTables[0] == "" {
		return "", fmt.Errorf("field %s not found for category %d", fieldName, categoryID)
	}

	specTable := specTables[0]
	specTableCache[key] = specTable
	return specTable, nil
}

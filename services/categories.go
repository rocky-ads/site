package services

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/models"
)

type categoryFieldKey struct {
	categoryID int
	fieldName  string
}

var (
	categoryCache  = make(map[int]models.Category)
	specTableCache = make(map[categoryFieldKey]string)
)

func InitCategories() error {
	var categories []models.Category
	query := `SELECT COALESCE(json_group_array(json_object('id', id, 'name', name, 'image_file', image_file)), '[]') FROM categories`
	if err := db.QueryJSON(&categories, query); err != nil {
		return fmt.Errorf("loading categories: %w", err)
	}

	for i := range categories {
		categoryCache[categories[i].ID] = categories[i]
	}

	type specTableData struct {
		CategoryID int    `json:"category_id"`
		FieldName  string `json:"field_name"`
		SpecTable  string `json:"spec_table"`
	}
	var specTables []specTableData
	specQuery := `
		SELECT COALESCE(json_group_array(json_object(
			'category_id', category_id,
			'field_name', field_name,
			'spec_table', spec_table
		)), '[]')
		FROM (
			SELECT DISTINCT
				c.category_id,
				f.name AS field_name,
				MIN(c.spec_table) AS spec_table
			FROM chain_fields cf
			JOIN chains c ON cf.chain_id = c.id
			JOIN fields f ON cf.field_id = f.id
			WHERE c.spec_table IS NOT NULL AND c.spec_table != ''
			GROUP BY c.category_id, f.name
		)
	`
	if err := db.QueryJSON(&specTables, specQuery); err != nil {
		return fmt.Errorf("loading spec tables: %w", err)
	}

	for _, st := range specTables {
		key := categoryFieldKey{categoryID: st.CategoryID, fieldName: st.FieldName}
		specTableCache[key] = st.SpecTable
	}

	return nil
}

func GetCategories() ([]models.Category, error) {
	var categories []models.Category
	for _, category := range categoryCache {
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})
	return categories, nil
}

func GetCategoryByID(categoryID int) (models.Category, error) {
	category, ok := categoryCache[categoryID]
	if !ok {
		return models.Category{}, fmt.Errorf("category not found: %d", categoryID)
	}
	return category, nil
}

func GetCategoryNameByID(categoryID int) (string, error) {
	category, err := GetCategoryByID(categoryID)
	if err != nil {
		return "", err
	}
	return category.Name, nil
}

func GetCategoryImageFile(categoryID int) (string, error) {
	category, err := GetCategoryByID(categoryID)
	if err != nil {
		return "", err
	}
	return category.ImageFile, nil
}

func GetCategoryIDByName(name string) (int, error) {
	for _, category := range categoryCache {
		if category.Name == name {
			return category.ID, nil
		}
	}
	return 0, fmt.Errorf("category not found: %s", name)
}

func ValidateCategory(category string) (int, error) {
	if category == "" {
		// Return default category ID if category is empty
		defaultCategoryID, err := GetCategoryIDByName(config.DefaultAdCategoryName)
		if err != nil {
			return 0, fmt.Errorf("default category not found: %w", err)
		}
		return defaultCategoryID, nil
	}

	categoryID, err := strconv.Atoi(category)
	if err != nil {
		return 0, fmt.Errorf("invalid category ID: %w", err)
	}

	if _, err := GetCategoryByID(categoryID); err != nil {
		return 0, fmt.Errorf("category not found: %w", err)
	}

	return categoryID, nil
}

func GetSpecTable(categoryID int, fieldName string) (string, error) {
	key := categoryFieldKey{categoryID: categoryID, fieldName: fieldName}
	specTable, ok := specTableCache[key]
	if !ok {
		return "", fmt.Errorf("field %s not found for category %d", fieldName, categoryID)
	}
	return specTable, nil
}

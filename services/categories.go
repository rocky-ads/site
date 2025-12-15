package services

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/db"
)

type categoryFieldKey struct {
	categoryID int
	fieldName  string
}

var (
	categoryNameCache  = make(map[string]int)
	categoryIDCache    = make(map[int]string)
	categoryImageCache = make(map[int]string)
	specTableCache     = make(map[categoryFieldKey]string)
)

func getCategories() error {
	type categoryData struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		ImageFile string `json:"image_file"`
	}

	var categories []categoryData
	query := `SELECT COALESCE(json_group_array(json_object('id', id, 'name', name, 'image_file', image_file)), '[]') FROM categories`
	if err := db.QueryJSON(&categories, query); err != nil {
		return fmt.Errorf("loading categories: %w", err)
	}

	for _, cat := range categories {
		categoryNameCache[cat.Name] = cat.ID
		categoryIDCache[cat.ID] = cat.Name
		categoryImageCache[cat.ID] = cat.ImageFile
	}

	return nil
}

func GetCategoryIDByName(categoryName string) (int, error) {
	if categoryID, ok := categoryNameCache[categoryName]; ok {
		return categoryID, nil
	}

	if err := getCategories(); err != nil {
		return 0, fmt.Errorf("loading categories: %w", err)
	}

	if categoryID, ok := categoryNameCache[categoryName]; ok {
		return categoryID, nil
	}

	return 0, fmt.Errorf("category not found: %s", categoryName)
}

func GetCategoryNameByID(categoryID int) (string, error) {
	if categoryName, ok := categoryIDCache[categoryID]; ok {
		return categoryName, nil
	}

	if err := getCategories(); err != nil {
		return "", fmt.Errorf("loading categories: %w", err)
	}

	if categoryName, ok := categoryIDCache[categoryID]; ok {
		return categoryName, nil
	}

	return "", fmt.Errorf("category not found: %d", categoryID)
}

func GetCategoryImageFile(categoryID int) (string, error) {
	if imageFile, ok := categoryImageCache[categoryID]; ok {
		return imageFile, nil
	}

	if err := getCategories(); err != nil {
		return "", fmt.Errorf("loading categories: %w", err)
	}

	if imageFile, ok := categoryImageCache[categoryID]; ok {
		return imageFile, nil
	}

	return "", fmt.Errorf("category image not found: %d", categoryID)
}

func ValidateCategory(category string) (int, *fiber.Error) {
	categoryID, err := strconv.Atoi(category)
	if err != nil {
		categoryID, err = GetCategoryIDByName(config.DefaultAdCategoryName)
		if err != nil {
			return 0, fiber.NewError(fiber.StatusNotFound, err.Error())
		}
	}

	if _, err := GetCategoryNameByID(categoryID); err != nil {
		return 0, fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return categoryID, nil
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

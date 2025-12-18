package ad

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rocky-ads/site/db"
)

type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ImageFile string `json:"image_file"`
}

var (
	categories = make(map[int]Category)
)

func Init() error {
	query := `
		SELECT COALESCE(
			json_group_array(
				json_object(
					'id', id,
					'name', name,
					'image_file', image_file
				)
			),
			'[]'
		)
		FROM categories`
	var allCategories []Category
	if err := db.QueryJSON(&allCategories, query); err != nil {
		return fmt.Errorf("loading categories: %w", err)
	}
	for _, cat := range allCategories {
		categories[cat.ID] = cat
	}
	return nil
}

func ParseCategory(categoryStr string) (int, error) {
	categoryID, err := strconv.Atoi(categoryStr)
	if err != nil {
		return 0, err
	}
	_, err = GetCategoryName(categoryID)
	if err != nil {
		return 0, err
	}
	return categoryID, nil
}

func GetCategoryName(categoryID int) (string, error) {
	category, ok := categories[categoryID]
	if !ok {
		return "", fmt.Errorf("category not found: %d", categoryID)
	}
	return category.Name, nil
}

func GetCategoryImageFile(categoryID int) (string, error) {
	category, ok := categories[categoryID]
	if !ok {
		return "", fmt.Errorf("category not found: %d", categoryID)
	}
	return category.ImageFile, nil
}

func GetCategories() ([]Category, error) {
	keys := make([]int, 0, len(categories))
	for k := range categories {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	cats := make([]Category, len(keys))
	for i, k := range keys {
		cats[i] = categories[k]
	}
	return cats, nil
}

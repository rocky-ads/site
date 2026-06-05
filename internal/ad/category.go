package ad

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
)

type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ImageFile string `json:"image_file"`
	Flags     int    `json:"flags"`
}

const (
	CategoryFlagMileage = 1 << 0
	CategoryFlagHours   = 1 << 1
)

func (c Category) HasMileage() bool {
	return c.Flags&CategoryFlagMileage != 0
}

func (c Category) HasHours() bool {
	return c.Flags&CategoryFlagHours != 0
}

var (
	categories      = make(map[int]Category)
	defaultCategory int
)

func LoadCategories() error {
	query := `
		SELECT COALESCE(
			json_group_array(
				json_object(
					'id', id,
					'name', name,
					'image_file', image_file,
					'flags', flags
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

	// Precalculate default category ID
	var err error
	defaultCategory, err = GetCategoryIDByName(config.DefaultAdCategoryName)
	if err != nil {
		return fmt.Errorf("default category not found: %w", err)
	}

	return nil
}

func ParseCategory(categoryStr string) int {
	categoryID, err := strconv.Atoi(categoryStr)
	if err != nil {
		return defaultCategory
	}
	_, ok := categories[categoryID]
	if !ok {
		return defaultCategory
	}
	return categoryID
}

func GetCategoryName(categoryID int) (string, error) {
	category, ok := categories[categoryID]
	if !ok {
		return "", fmt.Errorf("category not found: %d", categoryID)
	}
	return category.Name, nil
}

func GetCategory(categoryID int) (Category, error) {
	category, ok := categories[categoryID]
	if !ok {
		return Category{}, fmt.Errorf("category not found: %d", categoryID)
	}
	return category, nil
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

// GetCategoryIDByName returns the ID of a category by its name
func GetCategoryIDByName(name string) (int, error) {
	for id, cat := range categories {
		if cat.Name == name {
			return id, nil
		}
	}
	return 0, fmt.Errorf("category not found: %s", name)
}

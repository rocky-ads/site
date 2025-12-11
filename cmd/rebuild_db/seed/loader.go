package seed

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/logger"
	"github.com/rocky-ads/site/models"
)

// FieldData represents a field from fields.json
type FieldData struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	IsSpecField bool   `json:"is_spec_field"`
}

// CategoryFiles stores file information for a category
type CategoryFiles struct {
	ID         int
	SeedAdFile string
	ChainFiles map[string]string // Maps spec_table to chain_file
}

var categoryFiles = make(map[string]CategoryFiles)

// LoadAll loads all seed data into the database
func LoadAll(includeTestAds bool) error {
	startTime := time.Now()
	if err := LoadFields(); err != nil {
		return fmt.Errorf("loading fields: %w", err)
	}
	logger.Info("LoadFields completed", "duration", time.Since(startTime))

	startTime = time.Now()
	if err := LoadCategories(); err != nil {
		return fmt.Errorf("loading categories: %w", err)
	}
	logger.Info("LoadCategories completed", "duration", time.Since(startTime))

	startTime = time.Now()
	if err := LoadChains(); err != nil {
		return fmt.Errorf("loading chains: %w", err)
	}
	logger.Info("LoadChains completed", "duration", time.Since(startTime))

	if includeTestAds {
		startTime = time.Now()
		if err := LoadAds(includeTestAds); err != nil {
			return fmt.Errorf("loading ads: %w", err)
		}
		logger.Info("LoadAds completed", "duration", time.Since(startTime))
	}
	return nil
}

// LoadFields loads fields.json into the fields table
func LoadFields() error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/fields.json")
	if err != nil {
		return err
	}

	var fields []FieldData
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	for _, f := range fields {
		_, err := db.Exec(
			"INSERT OR REPLACE INTO fields (name, display_name) VALUES (?, ?)",
			f.Name, f.DisplayName,
		)
		if err != nil {
			return fmt.Errorf("inserting field %s: %w", f.Name, err)
		}
	}

	return nil
}

// LoadCategories loads ad-category.json into categories, chains, and chain_fields tables
func LoadCategories() error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/ad-category.json")
	if err != nil {
		return err
	}

	var categories []models.Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return err
	}

	// Sort categories by name to ensure deterministic processing order
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	for _, cat := range categories {
		// Insert category
		result, err := db.Exec(
			"INSERT INTO categories (name, seed_ad_file) VALUES (?, ?)",
			cat.Name, cat.SeedAdFile,
		)
		if err != nil {
			return fmt.Errorf("inserting category %s: %w", cat.Name, err)
		}

		categoryID, _ := result.LastInsertId()

		// Store category file mappings for later use
		chainFiles := make(map[string]string)
		for _, chain := range cat.Chains {
			if chain.SpecTable != "" && chain.ChainFile != "" {
				chainFiles[chain.SpecTable] = chain.ChainFile
			}
		}

		categoryFiles[cat.Name] = CategoryFiles{
			ID:         int(categoryID),
			SeedAdFile: cat.SeedAdFile,
			ChainFiles: chainFiles,
		}

		// Process chains and fields
		for chainIndex, chain := range cat.Chains {
			// Insert chain
			chainResult, err := db.Exec(
				"INSERT INTO chains (category_id, spec_table, chain_file, chain_index) VALUES (?, ?, ?, ?)",
				categoryID, chain.SpecTable, chain.ChainFile, chainIndex,
			)
			if err != nil {
				return fmt.Errorf("inserting chain %d for category %s: %w", chainIndex, cat.Name, err)
			}

			chainID, _ := chainResult.LastInsertId()

			// Process fields within this chain
			for fieldIndex, field := range chain.Fields {
				// Get field ID
				var fieldID int
				err := db.QueryRow("SELECT id FROM fields WHERE name = ?", field.Name).Scan(&fieldID)
				if err != nil {
					return fmt.Errorf("finding field %s: %w", field.Name, err)
				}

				// Calculate next_in_chain: if not last field in chain, point to next field's order
				nextInChain := 0
				if fieldIndex < len(chain.Fields)-1 {
					nextInChain = fieldIndex + 1
				}

				// Insert chain field
				_, err = db.Exec(
					"INSERT INTO chain_fields (chain_id, field_id, next_in_chain, is_required, field_order) VALUES (?, ?, ?, ?, ?)",
					chainID, fieldID, nextInChain, field.IsRequired, fieldIndex,
				)
				if err != nil {
					return fmt.Errorf("inserting chain field %s: %w", field.Name, err)
				}
			}
		}
	}

	return nil
}

// LoadChains loads vehicle and part chain files into chain combinations tables
func LoadChains() error {
	// Sort category names to ensure deterministic processing order
	categoryNames := make([]string, 0, len(categoryFiles))
	for name := range categoryFiles {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	// Use categoryFiles map populated during LoadCategories
	for _, categoryName := range categoryNames {
		files := categoryFiles[categoryName]
		categoryID := files.ID

		// Load chains based on spec_table -> chain_file mapping
		for specTable, chainFile := range files.ChainFiles {
			if specTable == "spec_part" {
				if err := loadPartChain(categoryID, chainFile); err != nil {
					return fmt.Errorf("loading part chain for category %s: %w", categoryName, err)
				}
			} else {
				// Vehicle chains (spec_vehicle, spec_bicycle, spec_ag)
				if err := loadVehicleChain(categoryID, chainFile, specTable); err != nil {
					return fmt.Errorf("loading vehicle chain for category %s: %w", categoryName, err)
				}
			}
		}
	}

	return nil
}

// loadVehicleChain loads a nested vehicle chain JSON file into the specified spec table
// specTable determines which table to insert into (spec_bicycle, spec_ag, or spec_vehicle)
func loadVehicleChain(categoryID int, filename string, specTable string) error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/" + filename)
	if err != nil {
		return err
	}

	var chainData map[string]interface{}
	if err := json.Unmarshal(data, &chainData); err != nil {
		return err
	}

	type bicycleRow struct {
		make  string
		model string
	}
	type agRow struct {
		make  string
		year  string
		model string
	}
	type vehicleRow struct {
		make   string
		year   string
		model  string
		engine string
	}

	var bicycleRows []bicycleRow
	var agRows []agRow
	var vehicleRows []vehicleRow

	// Recursively traverse the nested structure and collect rows
	var traverse func(map[string]interface{}, []string, int)
	traverse = func(m map[string]interface{}, path []string, depth int) {
		for key, value := range m {
			currentPath := append(path, key)
			switch v := value.(type) {
			case map[string]interface{}:
				traverse(v, currentPath, depth+1)
			case []interface{}:
				make := ""
				year := ""
				model := ""

				if len(currentPath) > 0 {
					make = currentPath[0]
				}
				if len(currentPath) > 1 {
					year = currentPath[1]
				}
				if len(currentPath) > 2 {
					model = currentPath[2]
				}

				pathLen := len(currentPath)
				if specTable == "spec_bicycle" {
					if pathLen == 1 {
						for _, modelVal := range v {
							modelStr := fmt.Sprintf("%v", modelVal)
							bicycleRows = append(bicycleRows, bicycleRow{make: make, model: modelStr})
						}
					}
				} else if specTable == "spec_ag" {
					if pathLen == 2 {
						for _, modelVal := range v {
							modelStr := fmt.Sprintf("%v", modelVal)
							agRows = append(agRows, agRow{make: make, year: year, model: modelStr})
						}
					}
				} else if specTable == "spec_vehicle" {
					if pathLen == 3 {
						for _, engineVal := range v {
							engineStr := fmt.Sprintf("%v", engineVal)
							vehicleRows = append(vehicleRows, vehicleRow{make: make, year: year, model: model, engine: engineStr})
						}
					}
				}
			}
		}
	}

	traverse(chainData, []string{}, 1)

	// Batch insert using transaction
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if specTable == "spec_bicycle" {
		stmt, err := tx.Prepare("INSERT OR IGNORE INTO spec_bicycle (category_id, make, model) VALUES (?, ?, ?)")
		if err != nil {
			return fmt.Errorf("preparing statement: %w", err)
		}
		defer stmt.Close()
		for _, row := range bicycleRows {
			if _, err := stmt.Exec(categoryID, row.make, row.model); err != nil {
				return fmt.Errorf("inserting row: %w", err)
			}
		}
	} else if specTable == "spec_ag" {
		stmt, err := tx.Prepare("INSERT OR IGNORE INTO spec_ag (category_id, make, year, model) VALUES (?, ?, ?, ?)")
		if err != nil {
			return fmt.Errorf("preparing statement: %w", err)
		}
		defer stmt.Close()
		for _, row := range agRows {
			if _, err := stmt.Exec(categoryID, row.make, row.year, row.model); err != nil {
				return fmt.Errorf("inserting row: %w", err)
			}
		}
	} else if specTable == "spec_vehicle" {
		stmt, err := tx.Prepare("INSERT OR IGNORE INTO spec_vehicle (category_id, make, year, model, engine) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			return fmt.Errorf("preparing statement: %w", err)
		}
		defer stmt.Close()
		for _, row := range vehicleRows {
			if _, err := stmt.Exec(categoryID, row.make, row.year, row.model, row.engine); err != nil {
				return fmt.Errorf("inserting row: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// loadPartChain loads a flat part chain JSON file (part_category->part_subcategory)
func loadPartChain(categoryID int, filename string) error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/" + filename)
	if err != nil {
		return err
	}

	var chainData map[string][]string
	if err := json.Unmarshal(data, &chainData); err != nil {
		return err
	}

	// Collect all rows first
	type partRow struct {
		partCategory    string
		partSubcategory string
	}
	var rows []partRow
	for partCategory, subcategories := range chainData {
		for _, partSubcategory := range subcategories {
			rows = append(rows, partRow{partCategory: partCategory, partSubcategory: partSubcategory})
		}
	}

	// Batch insert using transaction
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO spec_part (category_id, part_category, part_subcategory) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.Exec(categoryID, row.partCategory, row.partSubcategory); err != nil {
			return fmt.Errorf("inserting row: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// LoadAds loads ad files into ads and ad_values tables
// If includeTestAds is false, ads are not loaded (only schema and spec data)
func LoadAds(includeTestAds bool) error {
	if !includeTestAds {
		// Skip loading ads - only load schema and spec data
		return nil
	}

	// Sort category names to ensure deterministic processing order
	categoryNames := make([]string, 0, len(categoryFiles))
	for name := range categoryFiles {
		categoryNames = append(categoryNames, name)
	}
	sort.Strings(categoryNames)

	// Use categoryFiles map populated during LoadCategories
	for _, categoryName := range categoryNames {
		files := categoryFiles[categoryName]
		categoryID := files.ID

		if err := loadAdsFromFile(categoryID, files.SeedAdFile); err != nil {
			return fmt.Errorf("loading ads from %s for category %s: %w", files.SeedAdFile, categoryName, err)
		}
	}

	return nil
}

// adJSON represents the flat JSON structure from ad-*.json files
type adJSON struct {
	ID              int                 `json:"id,omitempty"`
	CategoryID      int                 `json:"category_id,omitempty"`
	Make            string              `json:"make,omitempty"`
	Years           []string            `json:"years,omitempty"`
	Models          []string            `json:"models,omitempty"`
	Engines         []string            `json:"engines,omitempty"`
	PartCategory    string              `json:"part_category,omitempty"`
	PartSubcategory string              `json:"part_subcategory,omitempty"`
	Title           string              `json:"title"`
	Description     string              `json:"description,omitempty"`
	Price           float64             `json:"price"`
	CreatedAt       string              `json:"created_at"`
	UserID          int                 `json:"user_id"`
	ImageCount      int                 `json:"image_count"`
	Location        models.LocationData `json:"location"`
}

// convertAdJSON converts adJSON (flat structure) to Ad (with SpecValues map)
func convertAdJSON(aj adJSON) models.Ad {
	ad := models.Ad{
		ID:          aj.ID,
		CategoryID:  aj.CategoryID,
		Title:       aj.Title,
		Description: aj.Description,
		Price:       aj.Price,
		CreatedAt:   aj.CreatedAt,
		UserID:      aj.UserID,
		ImageCount:  aj.ImageCount,
		Location:    aj.Location,
	}

	// Build SpecValues map from flat fields
	ad.SpecValues = make(models.FieldValues)
	if aj.Make != "" {
		ad.SpecValues["make"] = []string{aj.Make}
	}
	if len(aj.Years) > 0 {
		ad.SpecValues["year"] = aj.Years
	}
	if len(aj.Models) > 0 {
		ad.SpecValues["model"] = aj.Models
	}
	if len(aj.Engines) > 0 {
		ad.SpecValues["engine"] = aj.Engines
	}
	if aj.PartCategory != "" {
		ad.SpecValues["part_category"] = []string{aj.PartCategory}
	}
	if aj.PartSubcategory != "" {
		ad.SpecValues["part_subcategory"] = []string{aj.PartSubcategory}
	}

	return ad
}

// loadAdsFromFile loads ads from a single ad file
func loadAdsFromFile(categoryID int, filename string) error {
	data, err := os.ReadFile("cmd/rebuild_db/seed/" + filename)
	if err != nil {
		return err
	}

	var adsJSON []adJSON
	if err := json.Unmarshal(data, &adsJSON); err != nil {
		return err
	}

	// Convert adJSON to Ad
	ads := make([]models.Ad, len(adsJSON))
	for i, aj := range adsJSON {
		ads[i] = convertAdJSON(aj)
	}

	// Get field IDs for common fields
	fieldIDs := make(map[string]int)
	fieldNames := []string{"make", "year", "model", "engine", "part_category", "part_subcategory"}
	for _, name := range fieldNames {
		var fieldID int
		err := db.QueryRow("SELECT id FROM fields WHERE name = ?", name).Scan(&fieldID)
		if err == nil {
			fieldIDs[name] = fieldID
		}
	}

	for _, ad := range ads {
		// Serialize location as JSON
		locationJSON, _ := json.Marshal(ad.Location)

		// Insert ad
		result, err := db.Exec(
			"INSERT INTO ads (category_id, title, description, price, created_at, user_id, image_count, location) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			categoryID, ad.Title, ad.Description, ad.Price, ad.CreatedAt, ad.UserID, ad.ImageCount, string(locationJSON),
		)
		if err != nil {
			return fmt.Errorf("inserting ad: %w", err)
		}

		adID, _ := result.LastInsertId()

		// Insert spec field values from SpecValues map
		// All values go to ad_values table (one row per value)
		for fieldName, values := range ad.SpecValues {
			if len(values) == 0 {
				continue
			}

			fieldID, ok := fieldIDs[fieldName]
			if !ok {
				continue
			}

			// Insert all values (one row per value)
			for _, value := range values {
				if value != "" {
					db.Exec("INSERT OR IGNORE INTO ad_values (ad_id, field_id, value) VALUES (?, ?, ?)", adID, fieldID, value)
				}
			}
		}

	}

	return nil
}

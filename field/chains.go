package field

import (
	"fmt"

	"github.com/rocky-ads/site/db"
)

type CategoryChains struct {
	CategoryID   int
	CategoryName string
	Chains       []ChainGroup
}

type ChainGroup struct {
	ChainID    int
	ChainIndex int
	SpecTable  string
	Fields     []ChainField
}

type ChainField struct {
	ChainID     int
	ChainIndex  int
	SpecTable   string
	FieldOrder  int
	IsRequired  bool
	NextInChain int
	FieldID     int
	FieldName   string
	DisplayName string
	InputType   string
}

var categoryChainsCache = make(map[int]CategoryChains)

func Init() error {
	categoryChainsCache = make(map[int]CategoryChains)
	specFields = make(map[int]map[string]SpecField)

	var categoryIDs []int
	if err := db.Select(&categoryIDs, `SELECT id FROM categories ORDER BY id`); err != nil {
		return fmt.Errorf("loading categories: %w", err)
	}

	for _, categoryID := range categoryIDs {
		chains, err := loadCategoryChains(categoryID)
		if err != nil {
			return err
		}
		categoryChainsCache[categoryID] = chains
		initSpecFieldsForCategory(categoryID, chains)
	}

	return nil
}

func GetCategoryChainsMetadata(categoryID int) (CategoryChains, error) {
	chains, ok := categoryChainsCache[categoryID]
	if !ok {
		return CategoryChains{}, fmt.Errorf("category %d not found", categoryID)
	}
	return chains, nil
}

func FindChain(chains []ChainGroup, chainID int) (ChainGroup, bool) {
	for _, ch := range chains {
		if ch.ChainID == chainID {
			return ch, true
		}
	}
	return ChainGroup{}, false
}

func IsSpecChain(chain ChainGroup) bool {
	return chain.SpecTable != ""
}

func loadCategoryChains(categoryID int) (CategoryChains, error) {
	var cat struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	err := db.QueryRow(`SELECT id, name FROM categories WHERE id = ?`, categoryID).Scan(&cat.ID, &cat.Name)
	if err != nil {
		return CategoryChains{}, fmt.Errorf("category %d: %w", categoryID, err)
	}

	type row struct {
		ChainID     int    `db:"chain_id"`
		ChainIndex  int    `db:"chain_index"`
		SpecTable   string `db:"spec_table"`
		FieldOrder  int    `db:"field_order"`
		IsRequired  int    `db:"is_required"`
		NextInChain int    `db:"next_in_chain"`
		FieldID     int    `db:"field_id"`
		FieldName   string `db:"field_name"`
		DisplayName string `db:"display_name"`
		InputType   string `db:"input_type"`
	}

	var rows []row
	err = db.Select(&rows, `
		SELECT
			c.id AS chain_id,
			c.chain_index,
			COALESCE(c.spec_table, '') AS spec_table,
			cf.field_order,
			cf.is_required,
			cf.next_in_chain,
			f.id AS field_id,
			f.name AS field_name,
			f.display_name,
			f.input_type
		FROM chains c
		JOIN chain_fields cf ON cf.chain_id = c.id
		JOIN fields f ON f.id = cf.field_id
		WHERE c.category_id = ?
		ORDER BY c.chain_index, cf.field_order
	`, categoryID)
	if err != nil {
		return CategoryChains{}, err
	}

	chainIndex := make(map[int]int)
	var groups []ChainGroup
	for _, r := range rows {
		cf := ChainField{
			ChainID:     r.ChainID,
			ChainIndex:  r.ChainIndex,
			SpecTable:   r.SpecTable,
			FieldOrder:  r.FieldOrder,
			IsRequired:  r.IsRequired != 0,
			NextInChain: r.NextInChain,
			FieldID:     r.FieldID,
			FieldName:   r.FieldName,
			DisplayName: r.DisplayName,
			InputType:   r.InputType,
		}
		idx, ok := chainIndex[r.ChainID]
		if !ok {
			groups = append(groups, ChainGroup{
				ChainID:    r.ChainID,
				ChainIndex: r.ChainIndex,
				SpecTable:  r.SpecTable,
				Fields:     []ChainField{cf},
			})
			chainIndex[r.ChainID] = len(groups) - 1
			continue
		}
		groups[idx].Fields = append(groups[idx].Fields, cf)
	}

	return CategoryChains{
		CategoryID:   cat.ID,
		CategoryName: cat.Name,
		Chains:       groups,
	}, nil
}

func initSpecFieldsForCategory(categoryID int, chains CategoryChains) {
	categoryMap := make(map[string]SpecField)
	var lastSpecFieldName string

	for _, chain := range chains.Chains {
		if !IsSpecChain(chain) {
			continue
		}
		for j, cf := range chain.Fields {
			sf := chainFieldToSpecField(categoryID, chain, cf, j == 0)
			categoryMap[cf.FieldName] = sf
			lastSpecFieldName = cf.FieldName
		}
	}

	if lastSpecFieldName != "" {
		sf := categoryMap[lastSpecFieldName]
		sf.IsLastOverall = true
		categoryMap[lastSpecFieldName] = sf
	}

	if len(categoryMap) > 0 {
		specFields[categoryID] = categoryMap
	}
}

func chainFieldToSpecField(categoryID int, chain ChainGroup, cf ChainField, isFirst bool) SpecField {
	return SpecField{
		Field: Field{
			ID:          cf.FieldID,
			Name:        cf.FieldName,
			DisplayName: cf.DisplayName,
			InputType:   cf.InputType,
			CategoryID:  categoryID,
			IsRequired:  cf.IsRequired,
		},
		SpecTable: chain.SpecTable,
		IsFirst:   isFirst,
	}
}

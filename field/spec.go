package field

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rocky-ads/site/db"
)

// SpecField represents a specification field (a field in a spec table)
type SpecField struct {
	Field
	SpecTable     string `json:"spec_table"`
	IsLastOverall bool   `json:"is_last_overall"`
	IsFirst       bool   `json:"is_first"`
}

type SpecFielder interface {
	GetSpecField() SpecField
}

var (
	specFields = make(map[int]map[string]SpecField) // key: categoryID, then fieldName
)

func InitSpecFields() {
	for categoryID, fields := range categoryFields {
		categoryMap := make(map[string]SpecField)
		for _, f := range fields {
			if specFielder, ok := f.(SpecFielder); ok {
				specField := specFielder.GetSpecField()
				categoryMap[specField.Name] = specField
			}
		}
		if len(categoryMap) > 0 {
			specFields[categoryID] = categoryMap
		}
	}
}

var placeholderString string = strings.Repeat("?,", 1000)

func Placeholders(n int) string {
	if n == 0 {
		return ""
	}
	if n == 1 {
		return "?"
	}
	needed := 2*n - 1
	if needed <= len(placeholderString) {
		return placeholderString[:needed]
	}
	ph := make([]string, n)
	for i := range ph {
		ph[i] = "?"
	}
	return strings.Join(ph, ",")
}

// FilterSpecFields returns a new Values map containing only valid spec field names for the category.
// This filters out non-field parameters like "q" (search query) and "ad_ids".
func FilterSpecFields(categoryID int, fv Values) Values {
	fields, ok := categoryFields[categoryID]
	if !ok {
		return make(Values)
	}

	filtered := make(Values)
	for _, f := range fields {
		if _, ok := f.(SpecFielder); ok {
			fieldName := f.GetField().Name
			if values, exists := fv[fieldName]; exists {
				filtered[fieldName] = values
			}
		}
	}
	return filtered
}

func buildAdValuesQuery(f SpecField, fv Values, adFilterFunc func() (string, []any)) (string, []any, error) {

	adWhereClause, adArgs := adFilterFunc()

	query := `
		SELECT COALESCE(json_group_array(value), '[]') FROM (
			SELECT DISTINCT av.value FROM ad_values av
			JOIN fields f_main ON av.field_id = f_main.id
			WHERE av.ad_id IN (
				SELECT a.id FROM ads a
				WHERE ` + adWhereClause

	args := adArgs

	fv = FilterSpecFields(f.CategoryID, fv)

	for fieldName, values := range fv {
		if len(values) > 0 {
			query += fmt.Sprintf(` AND EXISTS (
				SELECT 1 FROM ad_values av_filter
				JOIN fields f_filter ON av_filter.field_id = f_filter.id
				WHERE av_filter.ad_id = a.id AND f_filter.name = ? AND av_filter.value IN (%s)
			)`, Placeholders(len(values)))
			args = append(args, fieldName)
			for _, v := range values {
				args = append(args, v)
			}
		}
	}

	query += `
			)
			AND f_main.name = ?
			ORDER BY av.value COLLATE NOCASE
		)
	`
	args = append(args, f.Name)

	return query, args, nil
}

func buildAllQuery(f SpecField, fv Values) (string, []any, error) {

	whereClauses := []string{"category_id = ?"}
	args := []any{f.CategoryID}

	fv = FilterSpecFields(f.CategoryID, fv)

	// For multi-value fields, we want the intersection: only return values of f.Name
	// that exist for ALL selected values of each filter field.
	// Single-value fields use a simple IN clause (equivalent behavior).
	var havingClauses []string

	for fieldName, values := range fv {
		if len(values) > 0 {
			whereClauses = append(whereClauses,
				fmt.Sprintf("%s IN (%s)", fieldName, Placeholders(len(values))))
			for _, v := range values {
				args = append(args, v)
			}
			if len(values) > 1 {
				havingClauses = append(havingClauses,
					fmt.Sprintf("COUNT(DISTINCT %s) = %d", fieldName, len(values)))
			}
		}
	}

	selectField := f.Name

	having := ""
	if len(havingClauses) > 0 {
		having = "GROUP BY " + selectField + " HAVING " + strings.Join(havingClauses, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(json_group_array(value), '[]')
		FROM (
			SELECT DISTINCT %s as value
			FROM %s
			WHERE %s
			%s
			ORDER BY %s COLLATE NOCASE
		)`, selectField, f.SpecTable, strings.Join(whereClauses, " AND "), having, selectField)

	return query, args, nil
}

func buildAnyQuery(f SpecField, fv Values) (string, []any, error) {
	return buildAdValuesQuery(f, fv, func() (string, []any) {
		return "a.category_id = ? AND a.deleted_at IS NULL", []any{f.CategoryID}
	})
}

func buildAdQuery(adIDs []int, f SpecField, fv Values) (query string, args []any, err error) {
	if len(adIDs) == 0 {
		return "SELECT '[]'", nil, nil
	}
	return buildAdValuesQuery(f, fv, func() (string, []any) {
		args := make([]any, len(adIDs))
		for i, id := range adIDs {
			args[i] = id
		}
		args = append(args, f.CategoryID)
		return "a.id IN (" + Placeholders(len(adIDs)) + ") AND a.category_id = ? AND a.deleted_at IS NULL", args
	})
}

func buildAndExecuteQuery(builder func() (string, []any, error)) (values []string, err error) {

	query, args, err := builder()
	if err != nil {
		return nil, err
	}

	err = db.QueryJSON(&values, query, args...)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (f SpecField) GetSpecField() SpecField {
	return f
}

func (f SpecField) GetAllValues(fv Values) ([]string, error) {
	return buildAndExecuteQuery(func() (string, []any, error) {
		return buildAllQuery(f, fv)
	})
}

func (f SpecField) GetAnyValues(fv Values) ([]string, error) {
	return buildAndExecuteQuery(func() (string, []any, error) {
		return buildAnyQuery(f, fv)
	})
}

func (f SpecField) GetAdValues(adIDs []int, fv Values) ([]string, error) {
	return buildAndExecuteQuery(func() (string, []any, error) {
		return buildAdQuery(adIDs, f, fv)
	})
}

func GetLastSpecField(categoryID int) (Fielder, error) {
	fields, err := GetFields(categoryID)
	if err != nil {
		return nil, err
	}

	// Find the last SpecFielder
	var lastSpecFielder Fielder
	for _, f := range fields {
		if _, ok := f.(SpecFielder); ok {
			lastSpecFielder = f
			// Continue to find the actual last one
		}
	}

	if lastSpecFielder == nil {
		return nil, fmt.Errorf("last spec field not found for category %d", categoryID)
	}

	return lastSpecFielder, nil
}

func GetFirstSpecFields(categoryID int) ([]Fielder, error) {
	fields, err := GetFields(categoryID)
	if err != nil {
		return nil, err
	}

	var specFields []Fielder
	for _, f := range fields {
		if specFielder, ok := f.(SpecFielder); ok {
			specField := specFielder.GetSpecField()
			if specField.IsFirst {
				specFields = append(specFields, f)
			}
		}
	}

	return specFields, nil
}

func GetSpecField(categoryID int, fieldName string) (SpecField, error) {
	categoryMap, ok := specFields[categoryID]
	if !ok {
		return SpecField{}, fmt.Errorf("fields not found for category %d", categoryID)
	}
	specField, ok := categoryMap[fieldName]
	if !ok {
		return SpecField{}, fmt.Errorf("field %s not found for category %d", fieldName, categoryID)
	}
	return specField, nil
}

// Chain represents a field chain within a category
type Chain struct {
	ChainIndex int            `json:"ChainIndex"`
	Fields     []FieldInChain `json:"Fields"`
}

// FieldInChain represents a field within a chain
type FieldInChain struct {
	Name        string `json:"Name"`
	DisplayName string `json:"DisplayName"`
	Order       int    `json:"Order"`
	NextInChain int    `json:"NextInChain"`
}

func GetCategoryChains(categoryID int) ([]Chain, error) {
	query := `
		SELECT 
			c.chain_index,
			cf.field_order,
			f.name,
			f.display_name,
			cf.next_in_chain
		FROM chains c
		JOIN chain_fields cf ON c.id = cf.chain_id
		JOIN fields f ON cf.field_id = f.id
		WHERE c.category_id = ? AND c.spec_table IS NOT NULL AND c.spec_table != ''
		ORDER BY c.chain_index, cf.field_order
	`

	type chainRow struct {
		ChainIndex  int    `db:"chain_index" json:"chain_index"`
		FieldOrder  int    `db:"field_order" json:"field_order"`
		Name        string `db:"name" json:"name"`
		DisplayName string `db:"display_name" json:"display_name"`
		NextInChain int    `db:"next_in_chain" json:"next_in_chain"`
	}

	var rows []chainRow
	if err := db.Select(&rows, query, categoryID); err != nil {
		return nil, fmt.Errorf("loading chains: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no chains found for category %d", categoryID)
	}

	// First pass: assign global Order values and build a map for lookup
	type fieldWithOrder struct {
		row         chainRow
		globalOrder int
	}
	fieldsWithOrder := make([]fieldWithOrder, len(rows))
	orderMap := make(map[string]int) // key: "chain_index:field_order", value: globalOrder

	globalOrder := 1
	for i, row := range rows {
		key := fmt.Sprintf("%d:%d", row.ChainIndex, row.FieldOrder)
		orderMap[key] = globalOrder
		fieldsWithOrder[i] = fieldWithOrder{row: row, globalOrder: globalOrder}
		globalOrder++
	}

	// Second pass: build chains with correct NextInChain values
	chainsMap := make(map[int][]FieldInChain)
	for _, fwo := range fieldsWithOrder {
		row := fwo.row
		nextInChain := 0
		if row.NextInChain > 0 {
			// Find the next field in the same chain and get its global Order
			nextKey := fmt.Sprintf("%d:%d", row.ChainIndex, row.NextInChain)
			if order, ok := orderMap[nextKey]; ok {
				nextInChain = order
			}
		}
		field := FieldInChain{
			Name:        row.Name,
			DisplayName: row.DisplayName,
			Order:       fwo.globalOrder,
			NextInChain: nextInChain,
		}
		chainsMap[row.ChainIndex] = append(chainsMap[row.ChainIndex], field)
	}

	// Get sorted chain indices
	var sortedIndices []int
	for chainIndex := range chainsMap {
		sortedIndices = append(sortedIndices, chainIndex)
	}
	sort.Ints(sortedIndices)

	// Convert to slice with renumbered chain indices starting from 0
	var chains []Chain
	for newIndex, oldIndex := range sortedIndices {
		if fields, ok := chainsMap[oldIndex]; ok {
			chains = append(chains, Chain{
				ChainIndex: newIndex,
				Fields:     fields,
			})
		}
	}

	return chains, nil
}

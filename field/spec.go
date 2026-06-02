package field

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rocky-ads/site/db"
)

type SpecField struct {
	Field
	SpecTable     string `json:"spec_table"`
	IsLastOverall bool   `json:"is_last_overall"`
	IsFirst       bool   `json:"is_first"`
}

var (
	specFields = make(map[int]map[string]SpecField)
)

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

func FilterSpecFields(categoryID int, fv Values) Values {
	chains, err := GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return make(Values)
	}

	filtered := make(Values)
	for _, chain := range chains.Chains {
		if !IsSpecChain(chain) {
			continue
		}
		for _, cf := range chain.Fields {
			if values, exists := fv[cf.FieldName]; exists {
				filtered[cf.FieldName] = values
			}
		}
	}
	return filtered
}

func FilterSpecFieldsForTable(categoryID int, specTable string, fv Values) Values {
	chains, err := GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return make(Values)
	}

	filtered := make(Values)
	for _, chain := range chains.Chains {
		if chain.SpecTable != specTable {
			continue
		}
		for _, cf := range chain.Fields {
			if values, exists := fv[cf.FieldName]; exists {
				filtered[cf.FieldName] = values
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

	fv = FilterSpecFieldsForTable(f.CategoryID, f.SpecTable, fv)

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

func GetLastSpecField(categoryID int) (Field, error) {
	categoryMap, ok := specFields[categoryID]
	if !ok {
		return Field{}, fmt.Errorf("last spec field not found for category %d", categoryID)
	}
	for _, sf := range categoryMap {
		if sf.IsLastOverall {
			return sf.Field, nil
		}
	}
	return Field{}, fmt.Errorf("last spec field not found for category %d", categoryID)
}

func GetFirstSpecFields(categoryID int) ([]Field, error) {
	chains, err := GetCategoryChainsMetadata(categoryID)
	if err != nil {
		return nil, err
	}

	var fields []Field
	for _, chain := range chains.Chains {
		if !IsSpecChain(chain) || len(chain.Fields) == 0 {
			continue
		}
		cf := chain.Fields[0]
		fields = append(fields, Field{
			ID:          cf.FieldID,
			Name:        cf.FieldName,
			DisplayName: cf.DisplayName,
			InputType:   cf.InputType,
			CategoryID:  categoryID,
			IsRequired:  cf.IsRequired,
		})
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("fields not found for category %d", categoryID)
	}
	return fields, nil
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

type Chain struct {
	ChainIndex int            `json:"ChainIndex"`
	Fields     []FieldInChain `json:"Fields"`
}

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
		ChainIndex  int    `db:"chain_index"`
		FieldOrder  int    `db:"field_order"`
		Name        string `db:"name"`
		DisplayName string `db:"display_name"`
		NextInChain int    `db:"next_in_chain"`
	}

	var rows []chainRow
	if err := db.Select(&rows, query, categoryID); err != nil {
		return nil, fmt.Errorf("loading chains: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no chains found for category %d", categoryID)
	}

	type fieldWithOrder struct {
		row         chainRow
		globalOrder int
	}
	fieldsWithOrder := make([]fieldWithOrder, len(rows))
	orderMap := make(map[string]int)

	globalOrder := 1
	for i, row := range rows {
		key := fmt.Sprintf("%d:%d", row.ChainIndex, row.FieldOrder)
		orderMap[key] = globalOrder
		fieldsWithOrder[i] = fieldWithOrder{row: row, globalOrder: globalOrder}
		globalOrder++
	}

	chainsMap := make(map[int][]FieldInChain)
	for _, fwo := range fieldsWithOrder {
		row := fwo.row
		nextInChain := 0
		if row.NextInChain > 0 {
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

	var sortedIndices []int
	for chainIndex := range chainsMap {
		sortedIndices = append(sortedIndices, chainIndex)
	}
	sort.Ints(sortedIndices)

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

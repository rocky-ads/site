package field

import (
	"fmt"
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

	for fieldName, values := range fv {
		if len(values) > 0 {
			whereClauses = append(whereClauses,
				fmt.Sprintf("%s IN (%s)", fieldName, Placeholders(len(values))))
			for _, v := range values {
				args = append(args, v)
			}
		}
	}

	selectField := f.Name

	query := fmt.Sprintf(`
		SELECT COALESCE(json_group_array(value), '[]')
		FROM (
			SELECT DISTINCT %s as value
			FROM %s
			WHERE %s
			ORDER BY %s COLLATE NOCASE
		)`, selectField, f.SpecTable, strings.Join(whereClauses, " AND "), selectField)

	return query, args, nil
}

func buildAnyQuery(f SpecField, fv Values) (string, []any, error) {
	return buildAdValuesQuery(f, fv, func() (string, []any) {
		return "a.category_id = ?", []any{f.CategoryID}
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
		return "a.id IN (" + Placeholders(len(adIDs)) + ") AND a.category_id = ?", args
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

func GetLastSpecField(categoryID int) (Fielder, error) {
	fields, err := GetFields(categoryID)
	if err != nil {
		return nil, err
	}

	for _, f := range fields {
		if specFielder, ok := f.(SpecFielder); ok {
			specField := specFielder.GetSpecField()
			if specField.IsLastOverall {
				return f, nil
			}
		}
	}

	return nil, fmt.Errorf("last spec field not found for category %d", categoryID)
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

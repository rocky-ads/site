package field

import (
	"fmt"
	"strings"

	"github.com/rocky-ads/site/db"
)

// SpecField represents a specification field (a field in a spec table)
type SpecField struct {
	Field
	SpecTable string `json:"spec_table"`
}

var placeholderString string = strings.Repeat("?,", 1000)

func placeholders(n int) string {
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

func fv2fi(fv Values) ValuesByIDs {
	fi := make(ValuesByIDs)
	for fieldName, values := range fv {
		if len(values) > 0 {
			if field, err := GetField(fieldName); err == nil {
				fi[field.ID] = values
			}
		}
	}
	return fi
}

func buildAdValuesQuery(f SpecField, fv Values, adFilterFunc func() (string, []any)) (string, []any, error) {
	fi := fv2fi(fv)

	adWhereClause, adArgs := adFilterFunc()

	query := `
		SELECT COALESCE(json_group_array(value), '[]') FROM (
			SELECT DISTINCT av.value FROM ad_values av
			WHERE av.ad_id IN (
				SELECT a.id FROM ads a
				WHERE ` + adWhereClause

	args := adArgs

	for fieldID, values := range fi {
		query += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM ad_values av_filter
			WHERE av_filter.ad_id = a.id AND av_filter.field_id = ? AND av_filter.value IN (%s)
		)`, placeholders(len(values)))
		args = append(args, fieldID)
		for _, v := range values {
			args = append(args, v)
		}
	}

	query += `
			)
			AND av.field_id = ?
			ORDER BY av.value COLLATE NOCASE
		)
	`
	args = append(args, f.ID)

	return query, args, nil
}

func buildAllQuery(f SpecField, fv Values) (string, []any, error) {

	whereClauses := []string{"category_id = ?"}
	args := []any{f.CategoryID}

	for fieldName, values := range fv {
		if len(values) > 0 {
			whereClauses = append(whereClauses,
				fmt.Sprintf("%s IN (%s)", fieldName, placeholders(len(values))))
			for _, v := range values {
				args = append(args, v)
			}
		}
	}

	selectField := f.Name

	query := fmt.Sprintf(`SELECT COALESCE(json_group_array(value), '[]') FROM (
		SELECT DISTINCT %s as value FROM %s WHERE %s ORDER BY %s COLLATE NOCASE
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
		return "a.id IN (" + placeholders(len(adIDs)) + ") AND a.category_id = ?", args
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

func (f SpecField) getAllValues(fv Values) ([]string, error) {
	return buildAndExecuteQuery(func() (string, []any, error) {
		return buildAllQuery(f, fv)
	})
}

func (f SpecField) getAnyValues(fv Values) ([]string, error) {
	return buildAndExecuteQuery(func() (string, []any, error) {
		return buildAnyQuery(f, fv)
	})
}

func (f SpecField) getAdValues(adIDs []int, fv Values) ([]string, error) {
	return buildAndExecuteQuery(func() (string, []any, error) {
		return buildAdQuery(adIDs, f, fv)
	})
}

package field

import (
	"fmt"
	"strings"

	"github.com/rocky-ads/site/db"
)

type ParentValues map[string][]string

func (p ParentValues) Values(fieldName string) []string {
	if p == nil {
		return nil
	}
	return p[fieldName]
}

func ListSpecOptions(chain ChainGroup, fieldName string, categoryID int, parents ParentValues) ([]string, error) {
	if chain.SpecTable == "" {
		return nil, fmt.Errorf("chain %d has no spec table", chain.ChainID)
	}
	if err := validateSpecTable(chain.SpecTable); err != nil {
		return nil, err
	}

	query, args, err := buildSpecOptionsQuery(chain, fieldName, categoryID, parents)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		options = append(options, value)
	}
	return options, rows.Err()
}

func ListAdFilterOptions(categoryID int, chain ChainGroup, fieldName string, parents ParentValues) ([]string, error) {
	priors, err := priorsBefore(chain, fieldName)
	if err != nil {
		return nil, err
	}

	for _, f := range priors {
		if len(parents.Values(f.FieldName)) == 0 {
			return nil, fmt.Errorf("missing parent field %q", f.FieldName)
		}
	}

	query := `
		SELECT DISTINCT av.value
		FROM ad_values av
		JOIN fields f ON f.id = av.field_id
		JOIN ads a ON a.id = av.ad_id
		WHERE a.category_id = ?
			AND a.deleted_at IS NULL
			AND f.name = ?
	`
	args := []any{categoryID, fieldName}

	for _, f := range priors {
		vals := parents.Values(f.FieldName)
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM ad_values av_p
				JOIN fields f_p ON f_p.id = av_p.field_id
				WHERE av_p.ad_id = a.id
					AND f_p.name = ?
					AND av_p.value IN (%s)
			)`, Placeholders(len(vals)))
		args = append(args, f.FieldName)
		for _, v := range vals {
			args = append(args, v)
		}
	}

	query += ` ORDER BY av.value`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		options = append(options, value)
	}
	return options, rows.Err()
}

func validateSpecTable(name string) error {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
		name,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("unknown spec table %q", name)
	}
	return nil
}

func priorsBefore(chain ChainGroup, fieldName string) ([]ChainField, error) {
	var priors []ChainField
	found := false
	for _, f := range chain.Fields {
		if f.FieldName == fieldName {
			found = true
			break
		}
		priors = append(priors, f)
	}
	if !found {
		return nil, fmt.Errorf("field %q not in chain %d", fieldName, chain.ChainID)
	}
	return priors, nil
}

func buildSpecOptionsQuery(chain ChainGroup, fieldName string, categoryID int, parents ParentValues) (string, []any, error) {
	priors, err := priorsBefore(chain, fieldName)
	if err != nil {
		return "", nil, err
	}

	table := chain.SpecTable
	args := []any{categoryID}

	if len(priors) == 0 {
		q := fmt.Sprintf(
			`SELECT DISTINCT %s FROM %s WHERE category_id = ? ORDER BY %s`,
			fieldName, table, fieldName,
		)
		return q, args, nil
	}

	for _, f := range priors {
		if len(parents.Values(f.FieldName)) == 0 {
			return "", nil, fmt.Errorf("missing parent field %q", f.FieldName)
		}
	}

	q := fmt.Sprintf(`SELECT %s FROM %s WHERE category_id = ?`, fieldName, table)
	expected := 1
	needIntersection := len(priors) > 1

	for _, f := range priors {
		vals := parents.Values(f.FieldName)
		expected *= len(vals)
		if len(vals) > 1 {
			needIntersection = true
		}
		col := f.FieldName
		if len(vals) == 1 {
			q += fmt.Sprintf(` AND %s = ?`, col)
			args = append(args, vals[0])
			continue
		}
		q += fmt.Sprintf(` AND %s IN (%s)`, col, Placeholders(len(vals)))
		for _, v := range vals {
			args = append(args, v)
		}
	}

	if !needIntersection {
		q = strings.Replace(q, fmt.Sprintf("SELECT %s", fieldName), fmt.Sprintf("SELECT DISTINCT %s", fieldName), 1)
		q += fmt.Sprintf(` ORDER BY %s`, fieldName)
		return q, args, nil
	}

	keyExpr := priors[0].FieldName
	for i := 1; i < len(priors); i++ {
		keyExpr += fmt.Sprintf(` || '|' || %s`, priors[i].FieldName)
	}
	q += fmt.Sprintf(
		` GROUP BY %s HAVING COUNT(DISTINCT %s) = ? ORDER BY %s`,
		fieldName, keyExpr, fieldName,
	)
	args = append(args, expected)

	return q, args, nil
}

func ParentsReady(chain ChainGroup, fieldName string, parents ParentValues) bool {
	priors, err := priorsBefore(chain, fieldName)
	if err != nil {
		return false
	}
	for _, f := range priors {
		if len(parents.Values(f.FieldName)) == 0 {
			return false
		}
	}
	return true
}

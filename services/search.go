package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/models"
)

// Optimized to use JOINs instead of EXISTS subqueries for better performance with large datasets
func Search(fv models.FieldValues, categoryID int) ([]int, error) {

	fi := fv2fi(fv)

	if len(fi) == 0 {
		var adIDs []int
		var query = "SELECT COALESCE(json_group_array(id), '[]') FROM ads WHERE category_id = ?"
		var args = []interface{}{categoryID}
		err := db.QueryJSON(&adIDs, query, args...)
		return adIDs, err
	}

	// Avoid JOIN to fields table by using cached field_id for better performance
	var query = "SELECT DISTINCT a.id FROM ads a WHERE a.category_id = ?"
	var args = []interface{}{categoryID}

	for fieldID, values := range fi {
		query += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM ad_values av
			WHERE av.ad_id = a.id AND av.field_id = ? AND av.value IN (%s)
		)`, placeholders(len(values)))
		args = append(args, fieldID)
		for _, v := range values {
			args = append(args, v)
		}
	}

	query = fmt.Sprintf("SELECT COALESCE(json_group_array(id), '[]') FROM (%s)", query)
	var adIDs []int
	err := db.QueryJSON(&adIDs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w\nQuery: %s\nArgs: %v", err, query, args)
	}
	return adIDs, nil
}

func LoadFilterValues(adID int) (models.FieldValues, error) {
	query := `
		SELECT f.name, av.value
		FROM ad_values av
		JOIN fields f ON av.field_id = f.id
		WHERE av.ad_id = ?
		ORDER BY f.name, av.value
	`
	rows, err := db.Query(query, adID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fv := make(models.FieldValues)
	for rows.Next() {
		var fieldName, value string
		if err := rows.Scan(&fieldName, &value); err != nil {
			return nil, err
		}
		fv[fieldName] = append(fv[fieldName], value)
	}

	return fv, rows.Err()
}

func Paths(fv models.FieldValues) string {
	if len(fv) == 0 {
		return ""
	}

	keys := make([]string, 0, len(fv))
	for k := range fv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, key := range keys {
		values := fv[key]
		if len(values) > 0 {
			sortedVals := make([]string, len(values))
			copy(sortedVals, values)
			sort.Strings(sortedVals)
			parts = append(parts, fmt.Sprintf("%s=%s", key, strings.Join(sortedVals, ",")))
		}
	}

	return strings.Join(parts, "::")
}

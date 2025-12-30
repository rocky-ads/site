package search

import (
	"fmt"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/field"
)

func Search(categoryID, limit, offset int, fv field.Values) ([]int, error) {

	if len(fv) == 0 {
		var adIDs []int
		query := `
			SELECT COALESCE(json_group_array(id), '[]')
			FROM (
				SELECT id FROM ads
				WHERE category_id = ? AND deleted_at IS NULL
				LIMIT ? OFFSET ?
			) AS limited_ads`
		var args = []any{categoryID, limit, offset}
		err := db.QueryJSON(&adIDs, query, args...)
		return adIDs, err
	}

	query := `
		SELECT DISTINCT a.id
		FROM ads a
		WHERE a.category_id = ? AND a.deleted_at IS NULL`
	var args = []any{categoryID}

	for fieldName, values := range fv {
		if len(values) > 0 {
			query += fmt.Sprintf(` AND EXISTS (
				SELECT 1 FROM ad_values av_filter
				JOIN fields f_filter ON av_filter.field_id = f_filter.id
				WHERE av_filter.ad_id = a.id AND f_filter.name = ? AND av_filter.value IN (%s)
			)`, field.Placeholders(len(values)))
			args = append(args, fieldName)
			for _, v := range values {
				args = append(args, v)
			}
		}
	}

	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	query = fmt.Sprintf("SELECT COALESCE(json_group_array(id), '[]') FROM (%s) AS limited_ads", query)
	var adIDs []int
	err := db.QueryJSON(&adIDs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w\nQuery: %s\nArgs: %v", err, query, args)
	}
	return adIDs, nil
}

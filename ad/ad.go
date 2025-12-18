package ad

import (
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/field"
)

func LoadFieldValues(adID int) (field.Values, error) {
	type fieldValue struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	var pairs []fieldValue
	query := `
		SELECT COALESCE(json_group_array(json_object(
			'name', f.name,
			'value', av.value
		)), '[]')
		FROM ad_values av
		JOIN fields f ON av.field_id = f.id
		WHERE av.ad_id = ?
		ORDER BY f.name, av.value
	`
	err := db.QueryJSON(&pairs, query, adID)
	if err != nil {
		return nil, err
	}

	fv := make(field.Values)
	for _, p := range pairs {
		fv[p.Name] = append(fv[p.Name], p.Value)
	}

	return fv, nil
}

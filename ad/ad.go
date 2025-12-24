package ad

import (
	"time"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/field"
)

type Ad struct {
	// Core database fields
	ID          int        `db:"id" json:"id"`
	CategoryID  int        `db:"category_id" json:"category_id"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	Price       int        `db:"price" json:"price"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	UserID      int        `db:"user_id" json:"user_id"`
	ImageCount  int        `db:"image_count" json:"image_count"`
	LocationID  int        `db:"location_id" json:"location_id"`

	// Location fields from joins
	City      string `db:"city" json:"city"`
	AdminArea string `db:"admin_area" json:"admin_area"`
	Country   string `db:"country" json:"country"`

	// User-specific computed fields
	Bookmarked bool `db:"-" json:"bookmarked"`
}

func GetAds(ids []int) ([]Ad, error) {
	if len(ids) == 0 {
		return []Ad{}, nil
	}
	query := `
		SELECT
			a.id,
			a.category_id,
			a.title,
			a.description,
			a.price,
			a.created_at,
			a.deleted_at,
			a.user_id,
			a.image_count,
			a.location_id,
			l.city,
			l.admin_area,
			l.country
		FROM ads a
		LEFT JOIN locations l ON a.location_id = l.id
		WHERE a.id IN (` + field.Placeholders(len(ids)) + `)
	`
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	var ads []Ad
	err := db.Select(&ads, query, args...)
	if err != nil {
		return nil, err
	}
	return ads, nil
}

func GetAd(id int) (Ad, error) {
	ads, err := GetAds([]int{id})
	if err != nil {
		return Ad{}, err
	}
	return ads[0], nil
}

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

package ad

import (
	"time"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/field"
)

type Ad struct {
	// Core database fields
	ID          int        `json:"id"`
	CategoryID  int        `json:"category_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Price       int        `json:"price"`
	CreatedAt   time.Time  `json:"created_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	UserID      int        `json:"user_id"`
	ImageCount  int        `json:"image_count"`
	LocationID  int        `json:"location_id"`

	// Location fields from joins
	City      string `json:"city"`
	AdminArea string `json:"admin_area"`
	Country   string `json:"country"`

	// User-specific computed fields
	Bookmarked bool `json:"bookmarked"`
}

func GetAds(ids []int) ([]Ad, error) {
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
		LEFT JOIN location l ON a.location_id = l.id
		WHERE a.id IN (?)
	`
	var ads []Ad
	err := db.QueryJSON(&ads, query, ids)
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

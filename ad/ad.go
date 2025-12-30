package ad

import (
	"fmt"
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

	// Location fields from join
	City      string `db:"city" json:"city"`
	AdminArea string `db:"admin_area" json:"admin_area"`
	Country   string `db:"country" json:"country"`

	// Computed fields
	Bookmarked bool `db:"bookmarked" json:"bookmarked"`
}

func (a Ad) IsDeleted() bool {
	return a.DeletedAt != nil
}

func GetAds(userID int, ids []int, loc *time.Location) ([]Ad, error) {
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
			l.country,
			CASE WHEN b.user_id IS NOT NULL THEN 1 ELSE 0 END AS bookmarked
		FROM ads a
		LEFT JOIN locations l ON a.location_id = l.id
		LEFT JOIN bookmarks b ON a.id = b.ad_id AND b.user_id = ?
		WHERE a.id IN (` + field.Placeholders(len(ids)) + `)
	`
	args := make([]any, len(ids)+1)
	args[0] = userID
	for i, id := range ids {
		args[i+1] = id
	}
	var ads []Ad
	err := db.Select(&ads, query, args...)
	if err != nil {
		return nil, err
	}

	// Convert timestamps to local timezone
	for i := range ads {
		ads[i].CreatedAt = ads[i].CreatedAt.In(loc)
		if ads[i].DeletedAt != nil {
			converted := (*ads[i].DeletedAt).In(loc)
			ads[i].DeletedAt = &converted
		}
	}

	return ads, nil
}

func GetAd(userID int, id int, loc *time.Location) (Ad, error) {
	ads, err := GetAds(userID, []int{id}, loc)
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

// GetUserAdIDs returns ad IDs for a user based on filter type
// filterType: "bookmarked", "active", or "deleted"
func GetUserAdIDs(userID int, filterType string) ([]int, error) {
	var query string
	var args []any

	switch filterType {
	case "bookmarked":
		query = `
			SELECT COALESCE(json_group_array(a.id), '[]')
			FROM ads a
			JOIN bookmarks b ON a.id = b.ad_id
			WHERE b.user_id = ? AND a.deleted_at IS NULL
			ORDER BY a.created_at DESC
		`
		args = []any{userID}
	case "active":
		query = `
			SELECT COALESCE(json_group_array(id), '[]')
			FROM ads
			WHERE user_id = ? AND deleted_at IS NULL
			ORDER BY created_at DESC
		`
		args = []any{userID}
	case "deleted":
		query = `
			SELECT COALESCE(json_group_array(id), '[]')
			FROM ads
			WHERE user_id = ? AND deleted_at IS NOT NULL
			ORDER BY deleted_at DESC
		`
		args = []any{userID}
	default:
		return nil, fmt.Errorf("invalid filter type: %s", filterType)
	}

	var adIDs []int
	err := db.QueryJSON(&adIDs, query, args...)
	return adIDs, err
}

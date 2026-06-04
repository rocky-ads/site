package ad

import (
	"fmt"
	"time"

	"github.com/rocky-ads/site/db"
)

type Ad struct {
	// Core database fields
	ID          int        `db:"id"`
	CategoryID  int        `db:"category_id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	Price         int        `db:"price"`
	PriceCurrency string     `db:"price_currency"`
	CreatedAt     time.Time  `db:"created_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	UserID      int        `db:"user_id"`
	ImageCount  int        `db:"image_count"`
	LocationID  int        `db:"location_id"`

	// Location fields from join
	City      string `db:"city"`
	AdminArea string `db:"admin_area"`
	Country   string `db:"country"`

	// Computed fields
	Bookmarked bool `db:"bookmarked"`
	RockCount  int  `db:"rock_count"`
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
			a.price_currency,
			a.created_at,
			a.deleted_at,
			a.user_id,
			a.image_count,
			a.location_id,
			l.city,
			l.admin_area,
			l.country,
			CASE WHEN b.user_id IS NOT NULL THEN 1 ELSE 0 END AS bookmarked,
			COALESCE((
				SELECT COUNT(*)
				FROM conversations c
				WHERE c.ad_id = a.id AND c.egg_thrower_id IS NOT NULL AND c.egg_thrower_id = c.enquirer_id
			), 0) AS rock_count
		FROM ads a
		LEFT JOIN locations l ON a.location_id = l.id
		LEFT JOIN bookmarks b ON a.id = b.ad_id AND b.user_id = ?
		WHERE a.id IN (` + db.Placeholders(len(ids)) + `)
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
	if len(ads) == 0 {
		return Ad{}, fmt.Errorf("ad not found")
	}
	return ads[0], nil
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

// CountActiveAdsByUser returns the number of non-deleted ads for a user
func CountActiveAdsByUser(userID int) (int, error) {
	var n int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM ads WHERE user_id = ? AND deleted_at IS NULL",
		userID,
	).Scan(&n)
	return n, err
}

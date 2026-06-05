package ad

import (
	"fmt"
	"strings"
	"time"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/facet"
)

type Ad struct {
	// Core database fields
	ID          int        `db:"id"`
	CategoryID  int        `db:"category_id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	CreatedAt   time.Time  `db:"created_at"`
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

	// Category-specific facet values
	Facets map[string]facet.Value
}

func (a Ad) IsDeleted() bool {
	return a.DeletedAt != nil
}

func (a Ad) Location() string {
	if a.City == "" && a.AdminArea == "" && a.Country == "" {
		return ""
	}

	var locationText string
	if a.City != "" && a.AdminArea != "" {
		locationText = a.City + ", " + a.AdminArea
	} else if a.City != "" {
		locationText = a.City
	} else if a.AdminArea != "" {
		locationText = a.AdminArea
	}

	var flag string
	if len(a.Country) == 2 {
		code := strings.ToUpper(a.Country)
		flag = string(rune(int32(code[0])-'A'+0x1F1E6)) + string(rune(int32(code[1])-'A'+0x1F1E6))
	}

	return flag + " " + locationText
}

func (a Ad) facetDefs() []facet.Def {
	cat, err := GetCategory(a.CategoryID)
	if err != nil {
		return nil
	}
	return cat.Facets()
}

// PriceValue returns the ad's price amount and currency, if a price facet is set.
func (a Ad) PriceValue() (amount int, currency string, ok bool) {
	v, exists := a.Facets["price"]
	if !exists || v.Num == nil {
		return 0, "", false
	}
	if v.Text != nil {
		currency = *v.Text
	}
	return *v.Num, currency, true
}

// FacetLabel returns the compact, non-price facet labels for a listing card,
// joined with separators (e.g. "45K mi · 2020").
func (a Ad) FacetLabel() string {
	return strings.Join(a.facetLabels(true), " · ")
}

// FacetLabels returns the full, non-price facet labels for the ad detail page.
func (a Ad) FacetLabels() []string {
	return a.facetLabels(false)
}

func (a Ad) facetLabels(compact bool) []string {
	var labels []string
	for _, d := range a.facetDefs() {
		if d.Key == "price" {
			continue
		}
		v, ok := a.Facets[d.Key]
		if !ok {
			continue
		}
		var s string
		if compact {
			s = d.FormatCompact(v)
		} else {
			s = d.FormatFull(v)
		}
		if s != "" {
			labels = append(labels, s)
		}
	}
	return labels
}

var placeholderString = strings.Repeat("?,", 1000)

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
		WHERE a.id IN (` + placeholders(len(ids)) + `)
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

	if err := attachFacets(ads); err != nil {
		return nil, err
	}

	return ads, nil
}

type facetRow struct {
	AdID int     `db:"ad_id"`
	Key  string  `db:"key"`
	Num  *int    `db:"num"`
	Text *string `db:"text"`
}

func attachFacets(ads []Ad) error {
	if len(ads) == 0 {
		return nil
	}
	ids := make([]int, len(ads))
	index := make(map[int]int, len(ads))
	for i := range ads {
		ads[i].Facets = map[string]facet.Value{}
		ids[i] = ads[i].ID
		index[ads[i].ID] = i
	}

	query := `SELECT ad_id, "key", num, "text" FROM ad_facets WHERE ad_id IN (` +
		placeholders(len(ids)) + `)`
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	var rows []facetRow
	if err := db.Select(&rows, query, args...); err != nil {
		return err
	}
	for _, r := range rows {
		if i, ok := index[r.AdID]; ok {
			ads[i].Facets[r.Key] = facet.Value{Num: r.Num, Text: r.Text}
		}
	}
	return nil
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

func Delete(id int) error {
	_, err := db.Exec(
		"UPDATE ads SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	return err
}

func Restore(id int) error {
	_, err := db.Exec("UPDATE ads SET deleted_at = NULL WHERE id = ?", id)
	return err
}

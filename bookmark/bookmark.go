package bookmark

import "github.com/rocky-ads/site/db"

func Add(userID, adID int) error {
	_, err := db.Exec(
		`INSERT INTO bookmarks (user_id, ad_id)
		VALUES (?, ?)
		ON CONFLICT (user_id, ad_id)
		DO UPDATE SET bookmarked_at = CURRENT_TIMESTAMP`,
		userID, adID,
	)
	return err
}

func Remove(userID, adID int) error {
	_, err := db.Exec(
		"DELETE FROM bookmarks WHERE user_id = ? AND ad_id = ?",
		userID, adID,
	)
	return err
}

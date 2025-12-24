package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/ui"
)

func BookmarkHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	csrfToken := local.GetCSRFToken(c)

	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}

	var bookmarked bool
	var handlerErr error

	switch c.Method() {
	case fiber.MethodPost:
		bookmarked = true
		handlerErr = bookmarkPost(userID, adID)
	case fiber.MethodDelete:
		bookmarked = false
		handlerErr = bookmarkDelete(userID, adID)
	default:
		return fiber.NewError(fiber.StatusMethodNotAllowed, "Method not allowed")
	}

	if handlerErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, handlerErr.Error())
	}

	return render(c, ui.Bookmark(adID, bookmarked, csrfToken))
}

func bookmarkPost(userID int, adID int) error {
	_, err := db.Exec(
		`INSERT INTO bookmarks (user_id, ad_id)
		VALUES (?, ?)
		ON CONFLICT (user_id, ad_id)
		DO UPDATE SET bookmarked_at = CURRENT_TIMESTAMP`,
		userID, adID,
	)
	return err
}

func bookmarkDelete(userID int, adID int) error {
	_, err := db.Exec(
		"DELETE FROM bookmarks WHERE user_id = ? AND ad_id = ?",
		userID, adID,
	)
	return err
}

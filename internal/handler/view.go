package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/ui"
)

func ViewHandler(c *fiber.Ctx) error {
	userID := local.GetUserID(c)
	viewStr := c.Params("view")
	view := ui.ValidateView(viewStr)
	loc := cookie.GetLocation(c)
	categoryID := cookie.GetCategoryID(c)
	csrfToken := local.GetCSRFToken(c)

	p := parseSearchParams(c, categoryID)
	results, err := searchAndRenderAds(p, userID, view, loc, csrfToken)
	if err != nil {
		return err
	}

	cookie.SetView(c, view)

	return render(c, ui.SearchView(userID, view, results))
}

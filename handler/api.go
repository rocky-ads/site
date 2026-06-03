package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/cookie"
	"github.com/rocky-ads/site/local"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/ui"
)

func SwitchCategoryHandler(c *fiber.Ctx) error {
	categoryID := param.GetCategoryID(c)

	if _, err := ad.GetCategoryName(categoryID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	cookie.SetCategoryID(c, categoryID)

	redirect := c.Query("return")
	if redirect == "" || redirect[0] != '/' || (len(redirect) > 1 && redirect[1] == '/') {
		redirect = "/"
	}
	c.Set("HX-Redirect", redirect)
	return c.Send(nil)
}

func ShowFiltersHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)

	userID := local.GetUserID(c)
	view := cookie.GetView(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	p := parseSearchParams(c, categoryID)
	results, err := searchAndRenderAds(p, userID, view, loc, csrfToken)
	if err != nil {
		return err
	}

	return render(c, ui.SearchWidget(userID, view, c.Query("q"), true, parseSearchFilters(c), results))
}

func SearchPageHandler(c *fiber.Ctx) error {
	categoryID := cookie.GetCategoryID(c)

	userID := local.GetUserID(c)
	view := cookie.GetView(c)
	loc := cookie.GetLocation(c)
	csrfToken := local.GetCSRFToken(c)

	p := parseSearchParams(c, categoryID)
	results, err := searchAndRenderAds(p, userID, view, loc, csrfToken)
	if err != nil {
		return err
	}

	return render(c, ui.SearchResults(view, results))
}

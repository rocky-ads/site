package cookie

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
)

func GetCategoryID(c *fiber.Ctx) int {
	category := c.Cookies("category")
	categoryID := ad.ParseCategory(category)
	if category == "" {
		SetCategoryID(c, categoryID)
	}
	return categoryID
}

func SetCategoryID(c *fiber.Ctx, category int) {
	c.Cookie(&fiber.Cookie{
		Name:     "category",
		Value:    strconv.Itoa(category),
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HTTPOnly: true,
		Secure:   config.CookieSecure,
		Path:     "/",
		SameSite: "Strict",
	})
}

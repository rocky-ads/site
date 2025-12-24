package cookie

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/ui"
)

func GetView(c *fiber.Ctx) int {
	view := c.Cookies("view")
	return ui.ValidateView(view)
}

func SetView(c *fiber.Ctx, view int) {
	c.Cookie(&fiber.Cookie{
		Name:     "view",
		Value:    strconv.Itoa(view),
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HTTPOnly: true,
		Secure:   config.CookieSecure,
		Path:     "/",
		SameSite: "Strict",
	})
}

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

func SetJWT(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   config.CookieSecure,
		Path:     "/",
		SameSite: "Strict",
		MaxAge:   24 * 60 * 60, // 24 hours
	})
}

func ClearJWT(c *fiber.Ctx) {
	c.ClearCookie("auth_token")
}

func GetJWT(c *fiber.Ctx) string {
	return c.Cookies("auth_token")
}

// GetLocation gets the timezone location from cookie
func GetLocation(c *fiber.Ctx) *time.Location {
	timezone := c.Cookies("timezone")
	if timezone == "" {
		// Default to UTC if no timezone cookie is set
		return time.UTC
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Invalid timezone, default to UTC
		return time.UTC
	}
	return loc
}

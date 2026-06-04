package cookie

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/ui"
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
		return time.UTC
	}
	// URL decode the timezone value (e.g., America%2FLos_Angeles -> America/Los_Angeles)
	decodedTimezone, err := url.QueryUnescape(timezone)
	if err != nil {
		return time.UTC
	}
	loc, err := time.LoadLocation(decodedTimezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

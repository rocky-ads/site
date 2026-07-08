package cookie

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/config"
)

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
	// Match SetJWT attributes so the browser actually deletes the cookie.
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   config.CookieSecure,
		Path:     "/",
		SameSite: "Strict",
		MaxAge:   -1,
	})
}

func GetJWT(c *fiber.Ctx) string {
	return c.Cookies("auth_token")
}

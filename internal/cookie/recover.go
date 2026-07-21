package cookie

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/config"
)

const recoverSessionCookie = "recover_session"

func SetRecoverSession(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     recoverSessionCookie,
		Value:    token,
		HTTPOnly: true,
		Secure:   config.CookieSecure,
		Path:     "/",
		SameSite: "Strict",
		MaxAge:   int(config.RecoverySessionTTL.Seconds()),
	})
}

func ClearRecoverSession(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     recoverSessionCookie,
		Value:    "",
		HTTPOnly: true,
		Secure:   config.CookieSecure,
		Path:     "/",
		SameSite: "Strict",
		MaxAge:   -1,
	})
}

func GetRecoverSession(c *fiber.Ctx) string {
	return c.Cookies(recoverSessionCookie)
}

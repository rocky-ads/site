package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/rocky-ads/site/internal/config"
)

// Registration and recovery limiters are built in InitRateLimiters so they
// share Redis storage from kv.Init.
var (
	RegistrationRateLimiter fiber.Handler
	RecoveryRateLimiter     fiber.Handler
)

// InitRateLimiters wires IP rate limiters. Call after kv.Init.
func InitRateLimiters(store fiber.Storage) {
	RegistrationRateLimiter = limiter.New(limiter.Config{
		Max:        config.EffectiveRegistrationRateLimitMax(),
		Expiration: config.RegistrationRateLimitExp,
		Storage:    store,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "reg:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			minutes := int(config.RegistrationRateLimitExp.Minutes())
			errorMsg := fmt.Sprintf("Too many registration attempts. "+
				"Please try again in %d minutes.", minutes)
			return showError(c, errorMsg)
		},
	})

	RecoveryRateLimiter = limiter.New(limiter.Config{
		Max:        config.EffectiveRecoveryRateLimitMax(),
		Expiration: config.RecoveryRateLimitExp,
		Storage:    store,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "rec:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			minutes := int(config.RecoveryRateLimitExp.Minutes())
			errorMsg := fmt.Sprintf("Too many recovery attempts. "+
				"Please try again in %d minutes.", minutes)
			return fiber.NewError(fiber.StatusTooManyRequests, errorMsg)
		},
	})
}

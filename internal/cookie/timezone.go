package cookie

import (
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetTimezone(c *fiber.Ctx) *time.Location {
	timezone := c.Cookies("timezone")
	if timezone == "" {
		return time.UTC
	}
	decodedTimezone, err := url.QueryUnescape(timezone)
	if err != nil {
		return time.UTC
	}
	tz, err := time.LoadLocation(decodedTimezone)
	if err != nil {
		return time.UTC
	}
	return tz
}

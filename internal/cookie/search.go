package cookie

import (
	"encoding/base64"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/facet"
)

const searchCookieName = "search"

// SearchState holds persisted search filters and panel expand/collapse.
type SearchState struct {
	Q        string                  `json:"q,omitempty"`
	Facets   map[string]facet.Filter `json:"facets,omitempty"`
	Location string                  `json:"location,omitempty"`
	Radius   int                     `json:"radius,omitempty"`
	Expanded bool                    `json:"expanded,omitempty"`
}

func GetSearchState(c *fiber.Ctx) SearchState {
	raw := c.Cookies(searchCookieName)
	if raw == "" {
		return SearchState{}
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return SearchState{}
	}
	var state SearchState
	if err := json.Unmarshal(data, &state); err != nil {
		return SearchState{}
	}
	return state
}

func SetSearchState(c *fiber.Ctx, state SearchState) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}
	c.Cookie(&fiber.Cookie{
		Name:     searchCookieName,
		Value:    base64.RawURLEncoding.EncodeToString(data),
		MaxAge:   30 * 24 * 60 * 60,
		HTTPOnly: true,
		Secure:   config.CookieSecure,
		Path:     "/",
		SameSite: "Strict",
	})
}

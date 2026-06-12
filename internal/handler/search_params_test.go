package handler

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/facet"
)

func TestParseFacetFiltersConditionCheckboxes(t *testing.T) {
	category := ad.Category{FacetKeys: []string{"condition"}}

	app := fiber.New()
	var filters map[string]facet.Filter
	app.Get("/search", func(c *fiber.Ctx) error {
		filters = parseFacetFilters(c, category)
		return c.SendStatus(fiber.StatusOK)
	})

	q := url.Values{}
	q.Add("condition", "New")
	q.Add("condition", "Used - Good")
	req := httptest.NewRequest("GET", "/search?"+q.Encode(), nil)
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
	got := filters["condition"]
	if len(got.Values) != 2 || got.Values[0] != "New" || got.Values[1] != "Used - Good" {
		t.Fatalf("condition filter = %+v", got)
	}
}

func TestParseFacetFiltersSaleWeek(t *testing.T) {
	category := ad.Category{FacetKeys: []string{"sale_start_date"}}

	app := fiber.New()
	var filters map[string]facet.Filter
	app.Get("/search", func(c *fiber.Ctx) error {
		filters = parseFacetFilters(c, category)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/search?sale_start_date=Next+week", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}
	got := filters["sale_start_date"]
	if got.Value == nil || *got.Value != "Next week" {
		t.Fatalf("Value = %v", got.Value)
	}
}

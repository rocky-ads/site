package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/field"
)

func parseAdFilters(c *fiber.Ctx) field.AdFilters {
	filters := field.AdFilters{Fields: make(map[string][]string)}

	if v := strings.TrimSpace(c.Query("price_min")); v != "" {
		if amount, err := strconv.Atoi(v); err == nil && amount >= 0 {
			filters.PriceMin = &amount
		}
	}
	if v := strings.TrimSpace(c.Query("price_max")); v != "" {
		if amount, err := strconv.Atoi(v); err == nil && amount >= 0 {
			filters.PriceMax = &amount
		}
	}

	filters.Location = strings.TrimSpace(c.Query("location"))

	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		name := string(key)
		if name == "category_id" || name == "price_min" || name == "price_max" || name == "location" || name == "q" || name == "page" {
			return
		}
		if field.IsFilterExcluded(name) {
			return
		}
		if len(value) == 0 {
			return
		}
		filters.Fields[name] = append(filters.Fields[name], string(value))
	})

	return filters
}

func parentValuesFromQuery(c *fiber.Ctx, fields []field.ChainField) field.ParentValues {
	parents := make(field.ParentValues)
	for _, f := range fields {
		var values []string
		c.Context().QueryArgs().VisitAll(func(key, value []byte) {
			if string(key) == f.FieldName && len(value) > 0 {
				values = append(values, string(value))
			}
		})
		if len(values) > 0 {
			parents[f.FieldName] = values
		}
	}
	return parents
}

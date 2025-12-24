package param

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/field"
)

func GetCategoryID(c *fiber.Ctx) int {
	category := c.Params("category")
	return ad.ParseCategory(category)
}

func GetSpecField(c *fiber.Ctx, categoryID int) (field.SpecField, error) {
	fieldName := c.Params("field")
	f, err := field.GetSpecField(categoryID, fieldName)
	if err != nil {
		return field.SpecField{}, err
	}
	return f, nil
}

func GetAdID(c *fiber.Ctx) (int, error) {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

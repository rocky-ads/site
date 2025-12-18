package param

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/field"
)

func GetCategoryID(c *fiber.Ctx) (int, error) {
	category := c.Params("category")
	return ad.ParseCategory(category)
}

func GetSpecField(c *fiber.Ctx, categoryID int) (field.SpecField, *fiber.Error) {
	fieldName := c.Params("field")
	f, err := field.GetSpecField(categoryID, fieldName)
	if err != nil {
		return field.SpecField{}, fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return f, nil
}

func GetAdID(c *fiber.Ctx) (int, *fiber.Error) {
	adIDStr := c.Params("id")
	adID, err := strconv.Atoi(adIDStr)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}
	return adID, nil
}

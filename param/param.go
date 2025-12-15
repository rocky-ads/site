package param

import (
	"strconv"

	"github.com/rocky-ads/site/models"
	"github.com/rocky-ads/site/services"

	"github.com/gofiber/fiber/v2"
)

func GetCategoryID(c *fiber.Ctx) (int, error) {
	category := c.Params("category")
	return services.ValidateCategory(category)
}

func GetSpecField(c *fiber.Ctx, categoryID int) (models.SpecField, *fiber.Error) {
	fieldName := c.Params("field")
	specField, err := services.GetSpecField(categoryID, fieldName)
	if err != nil {
		return models.SpecField{}, fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return specField, nil
}

func GetAdID(c *fiber.Ctx) (int, *fiber.Error) {
	adIDStr := c.Params("id")
	adID, err := strconv.Atoi(adIDStr)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Invalid ad ID")
	}
	return adID, nil
}

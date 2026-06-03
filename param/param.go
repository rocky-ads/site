package param

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/ad"
	"github.com/rocky-ads/site/config"
)

func GetCategoryID(c *fiber.Ctx) int {
	category := c.Params("category")
	return ad.ParseCategory(category)
}

func GetAdID(c *fiber.Ctx) (int, error) {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetPageLimitOffset(c *fiber.Ctx) (limit, offset int) {
	page := c.QueryInt("page", 1)
	limit = config.SearchPageSize
	offset = (page - 1) * limit
	return limit, offset
}

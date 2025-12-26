package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/param"
	"github.com/rocky-ads/site/ui"
)

func ImageHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	imageIDStr := c.Params("index")
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image ID")
	}

	size := c.Params("size")
	return renderSVG(c, ui.GenerateSVG(adID, imageID, size))
}

func GridImageHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	imageIDStr := c.Params("index")
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image ID")
	}

	countStr := c.Params("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image count")
	}

	return render(c, ui.GridImage(adID, count, imageID))
}

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

func ImageNavigationHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	imageID := c.QueryInt("index", 1)
	if imageID < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image index")
	}

	count := c.QueryInt("count", 1)
	if count < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image count")
	}

	size := c.Query("size", "480w")
	heightClass := c.Query("heightClass", "h-48")
	clickable := c.Query("clickable", "false") == "true"

	return render(c, ui.ImageNode(adID, count, imageID, size, heightClass, clickable))
}

func ImageFullScreenHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	imageID := c.QueryInt("index", 1)
	if imageID < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image index")
	}

	count := c.QueryInt("count", 1)
	if count < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image count")
	}

	size := c.Query("size", "1200w")

	// If update query parameter is present, it's a navigation update (use swap-oob)
	// Otherwise, it's an initial render (append to body)
	if c.Query("update") == "true" {
		return render(c, ui.ImageFullScreenUpdate(adID, imageID, count, size))
	}

	return render(c, ui.ImageFullScreen(adID, imageID, count, size))
}

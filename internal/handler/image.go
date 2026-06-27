package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/cookie"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/param"
	"github.com/rocky-ads/site/internal/ui"
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

	logger.Info("ImageHandler: fetching image",
		"adID", adID, "imageID", imageID, "size", size)

	data, err := adImageStore.Get(adID, imageID, size)
	if err == nil {
		logger.Info("ImageHandler: serving image",
			"adID", adID, "imageID", imageID, "size", size, "bytes", len(data))
		c.Set("Content-Type", "image/webp")
		return c.Send(data)
	}

	logger.Info("ImageHandler: image not found, falling back to SVG",
		"adID", adID, "imageID", imageID, "size", size, "error", err)
	return renderSVG(c, ui.GenerateSVG(adID, imageID, size))
}

func ImageNavigationHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	count, err := adImageCount(c, adID)
	if err != nil {
		return err
	}

	imageID := c.QueryInt("index", 1)
	if err := validateImageIndex(imageID, count); err != nil {
		return err
	}

	size := c.Query("size", "480w")
	heightClass := c.Query("heightClass", "h-48")
	clickable := c.Query("clickable", "false") == "true"
	userID := local.GetUserID(c)
	if clickable && userID != 0 {
		_ = ad.IncrementAdImageClickForUser(adID, userID, imageID)
	}

	if size == "1200w" && clickable {
		return render(c, ui.ImageNodeWithThumbnails(adID, count, imageID, size, heightClass, clickable))
	}

	return render(c, ui.ImageNode(adID, count, imageID, size, heightClass, clickable))
}

func ImageFullScreenHandler(c *fiber.Ctx) error {
	adID, err := param.GetAdID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	count, err := adImageCount(c, adID)
	if err != nil {
		return err
	}

	imageID := c.QueryInt("index", 1)
	if err := validateImageIndex(imageID, count); err != nil {
		return err
	}

	size := c.Query("size", "1200w")

	if c.Query("update") == "true" {
		return render(c, ui.ImageFullScreenUpdate(adID, imageID, count, size))
	}

	return render(c, ui.ImageFullScreen(adID, imageID, count, size))
}

func adImageCount(c *fiber.Ctx, adID int) (int, error) {
	userID := local.GetUserID(c)
	tz := cookie.GetTimezone(c)
	a, err := ad.GetAd(userID, adID, tz)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	if a.IsDeleted() && a.UserID != userID {
		return 0, fiber.NewError(fiber.StatusNotFound, "Ad not found")
	}
	return a.ImageCount, nil
}

func validateImageIndex(imageID, count int) error {
	if imageID < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image index")
	}
	maxIndex := count
	if maxIndex == 0 {
		maxIndex = 1
	}
	if imageID > maxIndex {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image index")
	}
	return nil
}

package handler

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Check if image file exists at static/images/ad/:id/:index-:size.webp
	imagePath := filepath.Join("static", "images", "ad", fmt.Sprintf("%d", adID), fmt.Sprintf("%d-%s.webp", imageID, size))
	logger.Info("ImageHandler: checking for file", "path", imagePath, "adID", adID, "imageID", imageID, "size", size)

	fileInfo, err := os.Stat(imagePath)
	if err == nil {
		// File exists, serve it
		logger.Info("ImageHandler: file found, serving", "path", imagePath, "size", fileInfo.Size())
		return c.SendFile(imagePath)
	}

	// File doesn't exist, fall back to SVG
	logger.Info("ImageHandler: file not found, falling back to SVG", "path", imagePath, "error", err)
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
	loc := cookie.GetLocation(c)
	a, err := ad.GetAd(userID, adID, loc)
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

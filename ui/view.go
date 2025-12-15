package ui

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// View types
const (
	ViewList int = iota + 1
	ViewGrid
	ViewTree
)

func ValidateView(viewStr string) (int, *fiber.Error) {
	view, err := strconv.Atoi(viewStr)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Invalid view")
	}
	if view < ViewList || view > ViewTree {
		return 0, fiber.NewError(fiber.StatusBadRequest, "Invalid view")
	}
	return view, nil
}

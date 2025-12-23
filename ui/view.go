package ui

import (
	"strconv"
)

// View types
const (
	ViewList int = iota + 1
	ViewGrid
	ViewTree
)

func ValidateView(viewStr string) int {
	view, err := strconv.Atoi(viewStr)
	if err != nil {
		return ViewGrid
	}
	if view < ViewList || view > ViewTree {
		return ViewGrid
	}
	return view
}

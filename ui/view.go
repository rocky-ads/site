package ui

import (
	"fmt"
	"strconv"
)

// View types
const (
	ViewList int = iota + 1
	ViewGrid
	ViewTree
)

func ValidateView(viewStr string) (int, error) {
	view, err := strconv.Atoi(viewStr)
	if err != nil {
		return 0, fmt.Errorf("invalid view: %w", err)
	}
	if view < ViewList || view > ViewTree {
		return 0, fmt.Errorf("invalid view: %d", view)
	}
	return view, nil
}

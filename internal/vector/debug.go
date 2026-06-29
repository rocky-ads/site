package vector

import (
	"fmt"
	"strings"
)

func formatDistances(rows []searchDistanceRow) string {
	if len(rows) == 0 {
		return "none"
	}
	parts := make([]string, len(rows))
	for i, r := range rows {
		parts[i] = fmt.Sprintf("ad=%d dist=%.4f", r.ID, r.Distance)
	}
	return strings.Join(parts, ", ")
}

type searchDistanceRow struct {
	ID       int
	Distance float64
}

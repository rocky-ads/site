package vector

import (
	"fmt"
	"math"
	"strings"
)

func truncateForLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func embeddingSummary(v []float32) (dim int, norm float64) {
	if len(v) == 0 {
		return 0, 0
	}
	dim = len(v)
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	return dim, math.Sqrt(sum)
}

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

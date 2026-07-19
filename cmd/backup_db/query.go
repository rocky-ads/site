package main

import (
	"fmt"
	"strings"

	"github.com/rocky-ads/site/internal/db"
)

func intInClause(ids []int) (string, []any) {
	if len(ids) == 0 {
		return "FALSE", nil
	}
	ph := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		ph[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	return strings.Join(ph, ", "), args
}

func uniqueInts(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// adsHasInactiveAt reports whether ads.inactive_at exists so backup/restore
// can work against DBs from before the ad lifecycle column was added.
func adsHasInactiveAt() (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'ads'
			  AND column_name = 'inactive_at'
		)
	`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check ads.inactive_at: %w", err)
	}
	return exists, nil
}

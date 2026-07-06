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

func columnExists(table, column string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = $1 AND column_name = $2
		)`, table, column).Scan(&exists)
	return exists, err
}

func tableExists(table string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`, table).Scan(&exists)
	return exists, err
}

// conversationThrowerColumns returns the thrower/thrown-at column names in
// conversations. Prefers rock_* (current schema); falls back to egg_* (legacy).
func conversationThrowerColumns() (throwerCol, thrownAtCol string, err error) {
	hasRock, err := columnExists("conversations", "rock_thrower_id")
	if err != nil {
		return "", "", err
	}
	if hasRock {
		return "rock_thrower_id", "rock_thrown_at", nil
	}
	hasEgg, err := columnExists("conversations", "egg_thrower_id")
	if err != nil {
		return "", "", err
	}
	if hasEgg {
		return "egg_thrower_id", "egg_thrown_at", nil
	}
	return "", "", fmt.Errorf("conversations table has neither rock_thrower_id nor egg_thrower_id")
}

// opinionsTableName returns the cached dispute opinions table (rock or legacy egg).
func opinionsTableName() (string, error) {
	hasRock, err := tableExists("rock_opinions")
	if err != nil {
		return "", err
	}
	if hasRock {
		return "rock_opinions", nil
	}
	hasEgg, err := tableExists("egg_opinions")
	if err != nil {
		return "", err
	}
	if hasEgg {
		return "egg_opinions", nil
	}
	return "", fmt.Errorf("neither rock_opinions nor egg_opinions table exists")
}

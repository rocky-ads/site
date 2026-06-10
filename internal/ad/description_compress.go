package ad

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/service/grok"
)

const (
	compressOriginalSystemPrompt = "You summarize classified ad descriptions. " +
		"Preserve key facts: price, condition, mechanical state, " +
		"features, and contact intent. Output only the summary text."
	compressHistorySystemPrompt = "You summarize classified ad edit history. " +
		"Preserve what changed and when it mattered. " +
		"Output only the summary text."
)

func compressWithGrok(systemPrompt, text string, maxRunes int) (string, error) {
	userPrompt := fmt.Sprintf(
		"Summarize to at most %d characters (Unicode runes). "+
			"Keep the most important facts.\n\n%s",
		maxRunes, text,
	)
	out, err := grok.CallGrok(systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(SanitizeAdText(out))
	return truncateRunes(out, maxRunes), nil
}

func descriptionRuneCount(desc string) int {
	return utf8.RuneCountInString(desc)
}

func truncateRunes(s string, max int) string {
	if descriptionRuneCount(s) <= max {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	n := 0
	for _, r := range s {
		if n >= max {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

func historyEntryText(e historyEntry) string {
	if e.body == "" {
		return DisplayDescription(e.header)
	}
	return DisplayDescription(e.header) + "\n\n" + e.body
}

// EnsureDescriptionFits compresses description text until it fits MaxAdDescriptionLength.
func EnsureDescriptionFits(
	desc string,
	at time.Time,
	loc *time.Location,
) (string, error) {
	return ensureDescriptionFits(desc, at, loc)
}

func ensureDescriptionFits(
	desc string,
	at time.Time,
	loc *time.Location,
) (string, error) {
	max := config.MaxAdDescriptionLength
	if descriptionRuneCount(desc) <= max {
		return desc, nil
	}

	original, history := SplitDescription(desc)
	entries := parseHistoryEntries(history)

	targetOriginal := max / 2
	if targetOriginal < 100 {
		targetOriginal = 100
	}
	compressed, err := compressWithGrok(
		compressOriginalSystemPrompt, original, targetOriginal,
	)
	if err != nil {
		return "", fmt.Errorf("compress description: %w", err)
	}
	original = compressed
	desc = assembleDescription(original, joinHistoryEntries(entries))
	desc = AppendHistoryEntry(
		desc,
		"Description compressed",
		"Original description was summarized to make room for edit history.",
		at,
		loc,
	)
	if descriptionRuneCount(desc) <= max {
		return desc, nil
	}

	original, history = SplitDescription(desc)
	entries = parseHistoryEntries(history)
	keepRecent := 2
	if len(entries) <= keepRecent {
		return truncateRunes(desc, max), nil
	}

	// History is newest-first; keep the newest entries, compress the rest.
	recentEntries := entries[:keepRecent]
	oldEntries := entries[keepRecent:]
	var oldParts []string
	for _, e := range oldEntries {
		oldParts = append(oldParts, historyEntryText(e))
	}
	oldText := strings.Join(oldParts, "\n\n")
	targetHistory := max / 4
	if targetHistory < 80 {
		targetHistory = 80
	}
	compressedHistory, err := compressWithGrok(
		compressHistorySystemPrompt, oldText, targetHistory,
	)
	if err != nil {
		return "", fmt.Errorf("compress history: %w", err)
	}
	mergedOld := historyEntry{
		header: historyMarker + formatHistoryTimestamp(at, loc) +
			"  History compressed",
		body: compressedHistory,
	}
	allEntries := append(recentEntries, mergedOld)
	desc = assembleDescription(original, joinHistoryEntries(allEntries))
	if descriptionRuneCount(desc) <= max {
		return desc, nil
	}
	return truncateRunes(desc, max), nil
}

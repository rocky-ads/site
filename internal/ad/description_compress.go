package ad

import (
	"fmt"
	"strings"
	"unicode/utf8"

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

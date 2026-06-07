package ad

import (
	"strings"
	"unicode"
)

// SanitizeAdText normalizes common paste artifacts and removes
// invisible or control characters unsuitable for stored ad title/description.
func SanitizeAdText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\n', '\r':
			b.WriteRune(r)
		case '\t':
			b.WriteByte(' ')
		default:
			if r == '\uFFFC' || r == '\uFFFD' {
				continue
			}
			if unicode.IsControl(r) || unicode.In(r, unicode.Cf) {
				continue
			}
			b.WriteRune(normalizeDescriptionRune(r))
		}
	}
	return b.String()
}

func normalizeDescriptionRune(r rune) rune {
	switch r {
	case '\u2018', '\u2019', '\u201A', '\u201B':
		return '\''
	case '\u201C', '\u201D', '\u201E', '\u201F':
		return '"'
	case '\u2013', '\u2014':
		return '-'
	case '\u00A0', '\u202F', '\u2007', '\u2008', '\u2009', '\u200A':
		return ' '
	default:
		return r
	}
}

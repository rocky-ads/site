package ad

import (
	"net/url"
	"regexp"
	"strings"
)

// DescriptionLink is a validated https URL rendered as link text.
type DescriptionLink struct {
	Text string
	URL  string
}

// DescriptionSegment is plain text or a link within a description.
type DescriptionSegment struct {
	Text string
	Link *DescriptionLink
}

var descriptionHTTPSURLPattern = regexp.MustCompile(`https://[^\s<>"']+`)

// SplitDescriptionLinks splits s into text and auto-detected https URL segments.
func SplitDescriptionLinks(s string) []DescriptionSegment {
	matches := descriptionHTTPSURLPattern.FindAllStringIndex(s, -1)
	if len(matches) == 0 {
		return []DescriptionSegment{{Text: s}}
	}
	segs := make([]DescriptionSegment, 0, len(matches)*2+1)
	last := 0
	for _, m := range matches {
		if m[0] > last {
			segs = append(segs, DescriptionSegment{Text: s[last:m[0]]})
		}
		rawMatch := s[m[0]:m[1]]
		href := trimURLTrailingPunctuation(rawMatch)
		if url, ok := validDescriptionHTTPSURL(href); ok {
			segs = append(segs, DescriptionSegment{
				Link: &DescriptionLink{Text: url, URL: url},
			})
			if tail := rawMatch[len(href):]; tail != "" {
				segs = append(segs, DescriptionSegment{Text: tail})
			}
		} else {
			segs = append(segs, DescriptionSegment{Text: rawMatch})
		}
		last = m[1]
	}
	if last < len(s) {
		segs = append(segs, DescriptionSegment{Text: s[last:]})
	}
	return segs
}

func trimURLTrailingPunctuation(raw string) string {
	for len(raw) > 0 {
		switch raw[len(raw)-1] {
		case '.', ',', ')', ']', '}', '!', '?', ';', ':', '\'', '"':
			raw = raw[:len(raw)-1]
		default:
			return raw
		}
	}
	return raw
}

func validDescriptionHTTPSURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", false
	}
	return raw, true
}

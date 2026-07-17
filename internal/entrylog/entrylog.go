package entrylog

import (
	"strings"
	"time"
	"unicode"
)

const (
	Marker = "\u001e"

	// Numeric offset so Parse is independent of time.Local (MST abbreviations
	// like PDT are only resolved when Local knows them; CI is usually UTC).
	TimestampLayout = "2006-01-02 03:04:05 PM -0700"
	legacyLayout    = "2006-01-02 03:04:05 PM MST"
)

// Fixed offsets for legacy journal/ad timestamps that used MST abbreviations.
var legacyZones = map[string]*time.Location{
	"UTC": time.UTC,
	"GMT": time.UTC,
	"PDT": time.FixedZone("PDT", -7*60*60),
	"PST": time.FixedZone("PST", -8*60*60),
	"MDT": time.FixedZone("MDT", -6*60*60),
	"MST": time.FixedZone("MST", -7*60*60),
	"CDT": time.FixedZone("CDT", -5*60*60),
	"CST": time.FixedZone("CST", -6*60*60),
	"EDT": time.FixedZone("EDT", -4*60*60),
	"EST": time.FixedZone("EST", -5*60*60),
}

// Block is one parsed marker-delimited entry.
type Block struct {
	Label string
	Meta  string
	Body  string
	At    time.Time
}

func SanitizeText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '\n', '\r':
			b.WriteRune(r)
		case '\t':
			b.WriteByte(' ')
		case '\u001e', '\u001f':
			continue
		default:
			if unicode.IsControl(r) || unicode.In(r, unicode.Cf) {
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

func FormatTimestamp(at time.Time, tz *time.Location) string {
	if tz != nil {
		at = at.In(tz)
	}
	return at.Format(TimestampLayout)
}

func ParseTimestamp(s string) (time.Time, error) {
	if t, err := time.Parse(TimestampLayout, s); err == nil {
		return t, nil
	}
	return parseLegacyTimestamp(s)
}

func parseLegacyTimestamp(s string) (time.Time, error) {
	abbr := s
	if i := strings.LastIndexByte(s, ' '); i >= 0 {
		abbr = s[i+1:]
	}
	loc, ok := legacyZones[abbr]
	if !ok {
		return time.Parse(legacyLayout, s)
	}
	return time.ParseInLocation(legacyLayout, s, loc)
}

func BuildBlock(label, meta, body string, at time.Time,
	tz *time.Location) string {
	body = strings.TrimSpace(SanitizeText(body))
	header := Marker + FormatTimestamp(at, tz) + "  " + label
	if meta != "" {
		header += "  " + meta
	}
	if body == "" {
		return header
	}
	return header + "\n\n" + body
}

func Parse(desc string) []Block {
	if desc == "" {
		return nil
	}
	parts := strings.Split(desc, Marker)
	var blocks []Block
	for _, part := range parts {
		part = strings.TrimLeft(part, "\n")
		if part == "" {
			continue
		}
		header, body, _ := strings.Cut(part, "\n\n")
		header = strings.TrimSpace(header)
		body = strings.TrimSpace(body)
		block, ok := parseHeader(header, body)
		if !ok {
			continue
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func parseHeader(header, body string) (Block, bool) {
	parts := strings.SplitN(header, "  ", 3)
	if len(parts) < 2 {
		return Block{}, false
	}
	at, err := ParseTimestamp(parts[0])
	if err != nil {
		return Block{}, false
	}
	meta := ""
	if len(parts) >= 3 {
		meta = parts[2]
	}
	return Block{
		Label: parts[1],
		Meta:  meta,
		Body:  body,
		At:    at,
	}, true
}

func Join(blocks []Block, tz *time.Location) string {
	if len(blocks) == 0 {
		return ""
	}
	var b strings.Builder
	for i, block := range blocks {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(BuildBlock(
			block.Label, block.Meta, block.Body, block.At, tz,
		))
	}
	return b.String()
}

// Append adds a block at the end (oldest-first timelines).
func Append(desc, label, meta, body string, at time.Time,
	tz *time.Location) string {
	block := BuildBlock(label, meta, body, at, tz)
	if desc == "" {
		return block
	}
	return strings.TrimRight(desc, "\n") + "\n\n" + block
}

// PrependAfterFirst inserts a block immediately after the first entry.
func PrependAfterFirst(desc, label, meta, body string, at time.Time,
	tz *time.Location) string {
	blocks := Parse(desc)
	newBlock := Block{
		Label: label,
		Meta:  meta,
		Body:  strings.TrimSpace(SanitizeText(body)),
		At:    at,
	}
	if tz != nil {
		newBlock.At = at.In(tz)
	}
	if len(blocks) == 0 {
		return BuildBlock(label, meta, body, at, tz)
	}
	out := []Block{blocks[0], newBlock}
	out = append(out, blocks[1:]...)
	return Join(out, tz)
}

func StripMarkers(s string) string {
	return strings.ReplaceAll(s, Marker, "")
}

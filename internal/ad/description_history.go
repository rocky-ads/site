package ad

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/location"
)

// historyMarker prefixes each edit-history entry header (stripped from user input).
const historyMarker = "\u001e"

// historyEndMarker separates immutable original text from edit history.
const historyEndMarker = "\u001f"

// HistoryEntryDisplay is one parsed edit-history block for UI rendering.
type HistoryEntryDisplay struct {
	Header       string
	Body         string
	ImageIndices []int
}

// DescriptionDisplay holds original ad text and edit history for display.
type DescriptionDisplay struct {
	Original string
	History  []HistoryEntryDisplay
}

// DisplayDescription strips server markers for plain-text rendering.
func DisplayDescription(desc string) string {
	s := strings.ReplaceAll(desc, historyEndMarker, "")
	return strings.ReplaceAll(s, historyMarker, "")
}

// ParseDescriptionForDisplay splits stored description into original and history.
func ParseDescriptionForDisplay(desc string) DescriptionDisplay {
	original, history := SplitDescription(desc)
	out := DescriptionDisplay{Original: original}
	for _, e := range parseHistoryEntries(history) {
		header := strings.TrimSpace(DisplayDescription(e.header))
		indices := imageIndicesFromHistoryEntry(header, e.body)
		body := e.body
		if len(indices) > 0 {
			body = ""
		}
		out.History = append(out.History, HistoryEntryDisplay{
			Header:       header,
			Body:         body,
			ImageIndices: indices,
		})
	}
	return out
}

// SplitDescription separates the immutable original body from server history.
func SplitDescription(desc string) (original, history string) {
	if i := strings.Index(desc, historyEndMarker); i >= 0 {
		return strings.TrimRight(desc[:i], "\n"),
			strings.TrimLeft(desc[i+len(historyEndMarker):], "\n")
	}
	// Legacy ads: history begins at the first entry marker.
	if i := strings.Index(desc, historyMarker); i >= 0 {
		return strings.TrimRight(desc[:i], "\n"),
			strings.TrimLeft(desc[i:], "\n")
	}
	return desc, ""
}

const imagesAddedLabel = "Images Added"

func formatImageAdditionBody(startIndex, count int) string {
	if count <= 0 {
		return ""
	}
	parts := make([]string, count)
	for i := range parts {
		parts[i] = strconv.Itoa(startIndex + i)
	}
	return strings.Join(parts, ",")
}

func imageIndicesFromHistoryEntry(header, body string) []int {
	if !strings.Contains(header, imagesAddedLabel) {
		return nil
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	var indices []int
	for _, part := range strings.Split(body, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 {
			continue
		}
		indices = append(indices, n)
	}
	return indices
}

func formatHistoryTimestamp(at time.Time, tz *time.Location) string {
	if tz != nil {
		at = at.In(tz)
	}
	return strings.ToLower(at.Format("1/2/2006 3:04 pm"))
}

func buildHistoryBlock(label, body string, at time.Time, tz *time.Location) string {
	body = strings.TrimSpace(SanitizeAdText(body))
	header := historyMarker + formatHistoryTimestamp(at, tz) +
		"  " + label
	if body == "" {
		return header
	}
	return header + "\n\n" + body
}

// AppendHistoryEntry prepends a labeled, timestamped history block (newest first).
func AppendHistoryEntry(
	desc, label, body string,
	at time.Time,
	tz *time.Location,
) string {
	body = strings.TrimSpace(SanitizeAdText(body))
	if body == "" && label != "Description compressed" &&
		label != "History compressed" {
		return desc
	}
	block := buildHistoryBlock(label, body, at, tz)
	original, history := SplitDescription(desc)
	original = strings.TrimRight(original, "\n")
	if history == "" {
		if original == "" {
			return historyEndMarker + block
		}
		return original + historyEndMarker + block
	}
	return original + historyEndMarker + block + "\n\n" + history
}

type historyEntry struct {
	header string
	body   string
}

func parseHistoryEntries(history string) []historyEntry {
	if history == "" {
		return nil
	}
	parts := strings.Split(history, historyMarker)
	var entries []historyEntry
	for _, part := range parts {
		part = strings.TrimLeft(part, "\n")
		if part == "" {
			continue
		}
		header, body, _ := strings.Cut(part, "\n\n")
		entries = append(entries, historyEntry{
			header: historyMarker + header,
			body:   body,
		})
	}
	return entries
}

func joinHistoryEntries(entries []historyEntry) string {
	if len(entries) == 0 {
		return ""
	}
	var b strings.Builder
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(e.header)
		if e.body != "" {
			b.WriteString("\n\n")
			b.WriteString(e.body)
		}
	}
	return b.String()
}

func assembleDescription(original, history string) string {
	original = strings.TrimRight(original, "\n")
	history = strings.TrimLeft(history, "\n")
	if original == "" && history == "" {
		return ""
	}
	if history == "" {
		return original
	}
	if original == "" {
		return historyEndMarker + history
	}
	return original + historyEndMarker + history
}

func facetValuesEqual(a, b facet.Value) bool {
	if a.Num == nil && b.Num != nil || a.Num != nil && b.Num == nil {
		return false
	}
	if a.Num != nil && b.Num != nil && *a.Num != *b.Num {
		return false
	}
	aVals := a.MultiEnumValues()
	bVals := b.MultiEnumValues()
	if len(aVals) > 0 || len(bVals) > 0 {
		if len(aVals) != len(bVals) {
			return false
		}
		for i := range aVals {
			if aVals[i] != bVals[i] {
				return false
			}
		}
		return true
	}
	if a.Text == nil && b.Text != nil || a.Text != nil && b.Text == nil {
		return false
	}
	if a.Text != nil && b.Text != nil && *a.Text != *b.Text {
		return false
	}
	return true
}

func formatPriceChange(old, new facet.Value) string {
	oldAmt, oldCur, oldOK := priceFromValue(old)
	newAmt, newCur, newOK := priceFromValue(new)
	if !oldOK && !newOK {
		return ""
	}
	if !oldOK && newOK {
		if newAmt == 0 {
			return "Listed as FREE"
		}
		return fmt.Sprintf("Price set to %s", currency.Format(newAmt, newCur))
	}
	if oldOK && !newOK {
		return fmt.Sprintf("Price removed (was %s)", currency.Format(oldAmt, oldCur))
	}
	if oldAmt == 0 && newAmt == 0 {
		return ""
	}
	if oldAmt == 0 {
		return fmt.Sprintf("Price set to %s", currency.Format(newAmt, newCur))
	}
	if newAmt == 0 {
		return fmt.Sprintf("Listed as FREE (was %s)", currency.Format(oldAmt, oldCur))
	}
	oldStr := currency.Format(oldAmt, oldCur)
	newStr := currency.Format(newAmt, newCur)
	if newAmt < oldAmt {
		return fmt.Sprintf("Price dropped from %s to %s", oldStr, newStr)
	}
	if newAmt > oldAmt {
		return fmt.Sprintf("Price increased from %s to %s", oldStr, newStr)
	}
	if oldCur != newCur {
		return fmt.Sprintf("Price currency changed from %s to %s", oldStr, newStr)
	}
	return ""
}

func priceFromValue(v facet.Value) (amount int, code string, ok bool) {
	if v.Num == nil {
		return 0, "", false
	}
	code = currency.Default
	if v.Text != nil {
		code = *v.Text
	}
	return *v.Num, code, true
}

func formatFacetChange(d facet.Def, old, new facet.Value) string {
	oldStr := d.FormatFull(old)
	newStr := d.FormatFull(new)
	if oldStr == "" && newStr == "" {
		return ""
	}
	if oldStr == "" {
		return fmt.Sprintf("%s set to %s", d.Label, newStr)
	}
	if newStr == "" {
		return fmt.Sprintf("%s removed (was %s)", d.Label, oldStr)
	}
	return fmt.Sprintf("%s changed from %s to %s", d.Label, oldStr, newStr)
}

func formatTitleChange(oldTitle, newTitle string) string {
	if oldTitle == newTitle {
		return ""
	}
	return fmt.Sprintf("Title changed from %q to %q", oldTitle, newTitle)
}

func formatLocationChange(oldText, newText string) string {
	oldText = strings.TrimSpace(oldText)
	newText = strings.TrimSpace(newText)
	if oldText == newText {
		return ""
	}
	if oldText == "" {
		return fmt.Sprintf("Location set to %s", newText)
	}
	if newText == "" {
		return fmt.Sprintf("Location removed (was %s)", oldText)
	}
	return fmt.Sprintf("Location changed from %s to %s", oldText, newText)
}

func formatLocationHistoryChange(
	oldAd Ad,
	newLocationText string,
	category Category,
) string {
	newRaw := strings.TrimSpace(newLocationText)
	if UsesFullAddressDisplay(category) {
		oldRaw := fullAddressText(oldAd)
		if strings.EqualFold(oldRaw, newRaw) {
			return ""
		}
		if newRaw == "" {
			return "Address removed"
		}
		return "Address changed"
	}

	oldRaw := strings.TrimSpace(oldAd.RawLocation)
	if strings.EqualFold(oldRaw, newRaw) {
		return ""
	}

	oldDisplay := location.DisplayText(
		oldAd.City, oldAd.AdminArea, oldAd.Country,
	)
	newDisplay := ""
	if newRaw != "" {
		display, _, _ := location.DisplayTextForInput(newRaw)
		newDisplay = display
	}
	return formatLocationChange(oldDisplay, newDisplay)
}

// BuildFieldChangeEntries returns history entry bodies for changed ad fields.
func BuildFieldChangeEntries(
	oldAd Ad,
	newTitle string,
	newLocationText string,
	newFacets map[string]facet.Value,
	category Category,
) []struct {
	label string
	body  string
} {
	var entries []struct {
		label string
		body  string
	}
	oldTitle := oldAd.Title
	newTitle = strings.TrimSpace(SanitizeAdText(newTitle))
	if body := formatTitleChange(oldTitle, newTitle); body != "" {
		entries = append(entries, struct {
			label string
			body  string
		}{"Title change", body})
	}
	if body := formatLocationHistoryChange(oldAd, newLocationText, category); body != "" {
		label := "Location change"
		if UsesFullAddressDisplay(category) {
			label = "Address change"
		}
		entries = append(entries, struct {
			label string
			body  string
		}{label, body})
	}
	for _, d := range category.Facets() {
		oldV, oldOK := oldAd.Facets[d.Key]
		newV, newOK := newFacets[d.Key]
		if !oldOK {
			oldV = facet.Value{}
		}
		if !newOK {
			newV = facet.Value{}
		}
		if facetValuesEqual(oldV, newV) {
			continue
		}
		var body string
		var label string
		if d.Key == "price" {
			body = formatPriceChange(oldV, newV)
			label = "Price change"
		} else if d.Kind == facet.Location {
			continue
		} else {
			body = formatFacetChange(d, oldV, newV)
			label = d.Label + " change"
		}
		if body != "" {
			entries = append(entries, struct {
				label string
				body  string
			}{label, body})
		}
	}
	return entries
}

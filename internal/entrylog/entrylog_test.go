package entrylog

import (
	"strings"
	"testing"
	"time"
)

func TestAppendAndParse(t *testing.T) {
	loc := time.UTC
	t1 := time.Date(2026, 7, 10, 22, 33, 0, 0, loc)
	t2 := time.Date(2026, 7, 10, 22, 35, 0, 0, loc)

	desc := BuildBlock("original", "", "hello", t1, loc)
	desc = Append(desc, "message", "sender:12", "world", t2, loc)

	blocks := Parse(desc)
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].Label != "original" || blocks[0].Body != "hello" {
		t.Errorf("first block: %+v", blocks[0])
	}
	if blocks[1].Label != "message" || blocks[1].Meta != "sender:12" {
		t.Errorf("second block: %+v", blocks[1])
	}
	if !blocks[0].At.Equal(t1) || !blocks[1].At.Equal(t2) {
		t.Errorf("times = %v %v", blocks[0].At, blocks[1].At)
	}
}

func TestParseLosAngelesRoundTrip(t *testing.T) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 7, 10, 22, 33, 0, 0, loc)
	got, err := ParseTimestamp(FormatTimestamp(at, loc))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(at) {
		t.Fatalf("got %v, want %v", got, at)
	}
}

func TestParseLegacyPDTAbbreviation(t *testing.T) {
	// Old rows used MST abbreviations; must not depend on time.Local.
	want := time.Date(2026, 7, 10, 22, 33, 0, 0,
		time.FixedZone("PDT", -7*60*60))
	got, err := ParseTimestamp("2026-07-10 10:33:00 PM PDT")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPrependAfterFirst(t *testing.T) {
	loc := time.UTC
	origAt := time.Date(2026, 1, 5, 12, 0, 0, 0, loc)
	editAt := time.Date(2026, 1, 7, 18, 47, 0, 0, loc)

	desc := BuildBlock("original", "", "Original text.", origAt, loc)
	desc = PrependAfterFirst(
		desc, "Price change", "", "Price dropped", editAt, loc,
	)

	blocks := Parse(desc)
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].Label != "original" {
		t.Fatalf("first = %q", blocks[0].Label)
	}
	if !strings.Contains(desc, blocks[1].Label) {
		t.Fatalf("second = %q", blocks[1].Label)
	}
}

func TestSanitizeStripsMarkers(t *testing.T) {
	got := SanitizeText("before\u001eafter\u001fend")
	if got != "beforeafterend" {
		t.Errorf("SanitizeText = %q", got)
	}
}

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

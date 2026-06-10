package ad

import (
	"testing"
	"time"
)

func TestEnsureDescriptionFitsNoOp(t *testing.T) {
	desc := "short description"
	got, err := EnsureDescriptionFits(desc, time.Now(), time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	if got != desc {
		t.Errorf("got %q", got)
	}
}

func TestTruncateRunes(t *testing.T) {
	got := truncateRunes("hello world", 5)
	if got != "hello" {
		t.Errorf("got %q", got)
	}
	if truncateRunes("hi", 5) != "hi" {
		t.Error("expected short string unchanged")
	}
}

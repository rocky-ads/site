package ui

import (
	"strings"
	"testing"
	"time"

	g "maragu.dev/gomponents"
)

func TestAdListNodeLocationStaysOnOneLine(t *testing.T) {
	html := renderAdNodes(t, []g.Node{
		AdListNode(0, 1, "$ 500", "57 Ford Fairlane convertible",
			"Yamhill County, OR", "1000K mi", true, time.Now(),
			true, false, "", false, 0, 0),
	})
	if strings.Contains(html, "max-w-[") {
		t.Fatal("list location should not cap width")
	}
	if !strings.Contains(html, "text-zinc-500 whitespace-nowrap") {
		t.Fatal("list location should stay on one line")
	}
	if !strings.Contains(html, "Yamhill County, OR") {
		t.Fatal("expected location text")
	}
	if !strings.Contains(html, "text-xs text-zinc-500 whitespace-nowrap") {
		t.Fatal("list mileage should stay on one line")
	}
	if !strings.Contains(html, " · 1000K mi") {
		t.Fatal("expected mileage text")
	}
}

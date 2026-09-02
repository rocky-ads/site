package ui

import (
	"strings"
	"testing"
)

func TestFunStatsPageGraphs(t *testing.T) {
	html := renderNodes(t, FunStatsPage(FunStatsData{
		Months: []FunStatsMonth{
			{
				Label:              "Jan 2026",
				RegisteredUsers:    2,
				UsersWithActiveAds: 1,
				ActiveAds:          1,
			},
			{
				Label:              "Feb 2026",
				RegisteredUsers:    3,
				UsersWithActiveAds: 2,
				ActiveAds:          4,
			},
		},
	}))
	for _, want := range []string{
		"Fun Stats",
		"Users and ads over time",
		"Registered users",
		"Users with active ads",
		"Active ads",
		"Jan 2026",
		"Feb 2026",
		"<svg",
		"<polyline",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("funstats missing %q", want)
		}
	}
	if strings.Count(html, "<svg") != 2 {
		t.Fatalf("want 2 graphs, got %d svgs", strings.Count(html, "<svg"))
	}
}

func TestFunStatsNiceScale(t *testing.T) {
	top, ticks := funStatsNiceScale(0)
	if top != 1 || len(ticks) < 2 || ticks[0] != 0 {
		t.Fatalf("zero: top=%d ticks=%v", top, ticks)
	}
	top, ticks = funStatsNiceScale(10)
	if top != 10 || ticks[0] != 0 || ticks[len(ticks)-1] != 10 {
		t.Fatalf("ten: top=%d ticks=%v", top, ticks)
	}
	for i := 1; i < len(ticks); i++ {
		if ticks[i] <= ticks[i-1] {
			t.Fatalf("ticks not increasing: %v", ticks)
		}
		if ticks[i]-ticks[i-1] != ticks[1]-ticks[0] {
			t.Fatalf("uneven ticks: %v", ticks)
		}
	}
}

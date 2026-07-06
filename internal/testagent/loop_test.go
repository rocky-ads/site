package testagent

import (
	"testing"

	"github.com/rocky-ads/site/internal/browserclient"
)

func TestModalLoopDetected(t *testing.T) {
	j := NewJournal()
	j.Append(JournalEntry{Phase: PhaseAct, Action: "GET /api/search-location-modal"})
	j.Append(JournalEntry{Phase: PhaseAct, Action: "GET /api/modal-remove/search-location"})
	j.Append(JournalEntry{Phase: PhaseAct, Action: "GET /api/search-location-modal"})
	if !modalLoopDetected(j.Entries()) {
		t.Fatal("expected modal loop detected")
	}
}

func TestEscapeModalLoop(t *testing.T) {
	page := browserclient.PageAffordances{
		Links: []browserclient.Link{{Href: "/faq", Label: "FAQ"}},
	}
	act := escapeModalLoop(page)
	if act.Path != "/faq" {
		t.Fatalf("got %q", act.Path)
	}
}

func TestNoopLoopDetected(t *testing.T) {
	j := NewJournal()
	j.Append(JournalEntry{Phase: PhasePlan, URL: "/auth/user/myads", Action: "noop"})
	j.Append(JournalEntry{Phase: PhasePlan, URL: "/auth/user/myads", Action: "noop"})
	j.Append(JournalEntry{Phase: PhasePlan, URL: "/auth/user/myads", Action: "noop"})
	if !noopLoopDetected(j.Entries(), "/auth/user/myads") {
		t.Fatal("expected noop loop")
	}
}

func TestMyAdsTabLoopDetected(t *testing.T) {
	j := NewJournal()
	for i := 0; i < 4; i++ {
		j.Append(JournalEntry{Phase: PhaseAct, Action: "CLICK /auth/user/myads/tab/active"})
	}
	if !myAdsTabLoopDetected(j.Entries()) {
		t.Fatal("expected myads tab loop")
	}
}

func TestEscapeNoopLoopCarSeller(t *testing.T) {
	act := escapeNoopLoop("/auth/user/myads", browserclient.PageAffordances{}, Persona{Name: "car_seller"})
	if act.Path != "/auth/ad/new" {
		t.Fatalf("got %q", act.Path)
	}
}

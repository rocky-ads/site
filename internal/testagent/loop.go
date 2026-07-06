package testagent

import (
	"strings"

	"github.com/rocky-ads/site/internal/browserclient"
)

func modalLoopDetected(entries []JournalEntry) bool {
	modalActs := 0
	start := len(entries) - 8
	if start < 0 {
		start = 0
	}
	for i := len(entries) - 1; i >= start; i-- {
		e := entries[i]
		if e.Phase != PhaseAct {
			continue
		}
		if browserclient.IsModalPath(pathFromAction(e.Action)) {
			modalActs++
		}
	}
	return modalActs >= 3
}

func noopLoopDetected(entries []JournalEntry, currentPath string) bool {
	count := 0
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Phase != PhasePlan || !strings.HasPrefix(e.Action, "noop") {
			continue
		}
		if e.URL != "" && e.URL != currentPath {
			continue
		}
		count++
		if count >= 3 {
			return true
		}
	}
	return false
}

func myAdsTabLoopDetected(entries []JournalEntry) bool {
	count := 0
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Phase != PhaseAct {
			continue
		}
		if strings.HasPrefix(e.Action, "CLICK /auth/user/myads/tab/") {
			count++
			if count >= 4 {
				return true
			}
		}
	}
	return false
}

func settingsTabLoopDetected(entries []JournalEntry) bool {
	count := 0
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Phase != PhaseAct {
			continue
		}
		if strings.HasPrefix(e.Action, "CLICK /auth/user/settings/") {
			count++
			if count >= 3 {
				return true
			}
		}
	}
	return false
}

func pathFromAction(action string) string {
	parts := strings.SplitN(strings.TrimSpace(action), " ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return action
}

func escapeModalLoop(page browserclient.PageAffordances) PlannedAction {
	for _, l := range page.Links {
		if l.Href == "/" {
			continue
		}
		return PlannedAction{
			Action: "get",
			Path:   l.Href,
			Reason: "escape modal loop",
		}
	}
	for _, c := range page.AdCards {
		return PlannedAction{
			Action: "get",
			Path:   "/ad/" + c.ID,
			Reason: "escape modal loop",
		}
	}
	for _, path := range []string{"/about", "/faq", "/auth/user/myads"} {
		return PlannedAction{
			Action: "get",
			Path:   path,
			Reason: "escape modal loop",
		}
	}
	return PlannedAction{Action: "get", Path: "/", Reason: "escape modal loop"}
}

func sellerFallbackAction(persona Persona, page browserclient.PageAffordances) PlannedAction {
	if persona.Name == "car_seller" {
		return PlannedAction{
			Action: "get",
			Path:   "/auth/ad/new",
			Reason: "already logged in; create an ad",
		}
	}
	return escapeStuckPage(page, persona)
}

func escapeStuckPage(page browserclient.PageAffordances, persona Persona) PlannedAction {
	for _, c := range page.AdCards {
		return PlannedAction{
			Action: "get",
			Path:   "/ad/" + c.ID,
			Reason: "escape stuck page",
		}
	}
	act := escapeModalLoop(page)
	act.Reason = "escape stuck page"
	return act
}

func escapeNoopLoop(path string, page browserclient.PageAffordances, persona Persona) PlannedAction {
	switch persona.Name {
	case "car_seller":
		if path != "/auth/ad/new" {
			return PlannedAction{
				Action: "get", Path: "/auth/ad/new",
				Reason: "escape noop loop",
			}
		}
	case "negotiator", "messenger", "cross_traffic", "bike_buyer":
		if path == "/auth/user/messages" {
			return PlannedAction{Action: "get", Path: "/", Reason: "escape noop loop"}
		}
	case "lurker", "newcomer", "power_searcher":
		for _, c := range page.AdCards {
			return PlannedAction{
				Action: "get", Path: "/ad/" + c.ID,
				Reason: "escape noop loop",
			}
		}
	}
	return escapeModalLoop(page)
}

func escapeMyAdsTabLoop(persona Persona, page browserclient.PageAffordances) PlannedAction {
	if persona.Name == "car_seller" {
		return PlannedAction{
			Action: "get", Path: "/auth/ad/new",
			Reason: "escape myads tab loop",
		}
	}
	for _, c := range page.AdCards {
		return PlannedAction{
			Action: "get", Path: "/ad/" + c.ID,
			Reason: "escape myads tab loop",
		}
	}
	return PlannedAction{Action: "get", Path: "/", Reason: "escape myads tab loop"}
}

func escapeSettingsTabLoop(page browserclient.PageAffordances, persona Persona) PlannedAction {
	_ = page
	_ = persona
	return PlannedAction{
		Action: "get", Path: "/auth/user/myads",
		Reason: "escape settings tab loop",
	}
}

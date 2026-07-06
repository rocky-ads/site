package testagent

import "fmt"

// Persona describes an agent role for Grok planning.
type Persona struct {
	Name        string
	Description string
}

// PreferredAdCategory returns the category name this persona should use when posting ads.
func (p Persona) PreferredAdCategory() string {
	switch p.Name {
	case "car_seller":
		return "Cars & Trucks"
	default:
		return ""
	}
}

// DefaultPersonas returns the standard ten agent personas.
func DefaultPersonas() []Persona {
	return []Persona{
		{Name: "newcomer", Description: "A first-time visitor exploring the site. Read FAQ and about pages, browse ads, and register or log in, create ads, delete ads, share ads, bookmark ads."},
		{Name: "car_seller", Description: "Someone who sells whole vehicles in Cars & Trucks (not Car & Truck Parts unless listing parts only). Create and manage ads at /auth/ad/new and /auth/user/myads."},
		{Name: "bike_buyer", Description: "A buyer looking for bicycles and bike parts. Browse ads via /ad/ links, use search filters, and log in to bookmark or save favorites."},
		{Name: "messenger", Description: "A buyer who finds ads and sends inquiries. Log in or register if required, then start and continue conversations."},
		{Name: "negotiator", Description: "An active buyer who messages sellers, negotiates, and may throw eggs in disputes. Authenticate when the site requires it."},
		{Name: "power_searcher", Description: "A user who tries many search queries, category switches, and location filters. May create an account to save preferences."},
		{Name: "settings_user", Description: "A user who manages their profile, notifications, and password at /auth/user/settings. Log in or register first if prompted."},
		{Name: "adversarial", Description: "A curious user who tries unusual inputs, empty fields, and edge cases while staying on allowed paths."},
		{Name: "lurker", Description: "A passive browser who mostly reads ads and occasionally bookmarks one. Log in only when bookmarking."},
		{Name: "cross_traffic", Description: "An engaged user who messages sellers and responds to conversations organically. Sign up or log in to participate."},
	}
}

// AgentPhone returns the E.164 phone for agent index (1-based).
func AgentPhone(index int) string {
	return fmtPhone(index)
}

// AgentUsername returns the username for agent index (1-based).
func AgentUsername(index int) string {
	return fmt.Sprintf("Agent%02d", index)
}

func fmtPhone(index int) string {
	return fmt.Sprintf("+1555010%04d", 1000+index)
}

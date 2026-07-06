package browserclient

import "strings"

// PageLinkPaths returns hrefs suitable for full-page navigation (get).
func PageLinkPaths(p PageAffordances) map[string]bool {
	htmx := HTMXPaths(p)
	out := map[string]bool{}
	for _, l := range p.Links {
		if IsBlockedAgentPath(l.Href) {
			continue
		}
		if htmx[l.Href] {
			continue
		}
		out[l.Href] = true
	}
	return out
}

// HTMXPaths returns in-page HTMX control targets (click).
func HTMXPaths(p PageAffordances) map[string]bool {
	out := map[string]bool{}
	for _, h := range p.HTMX {
		if IsBlockedAgentPath(h.Path) {
			continue
		}
		out[h.Path] = true
	}
	return out
}

// CategorySwitchPath finds an hx-get switch URL for the named category.
func CategorySwitchPath(p PageAffordances, name string) string {
	want := strings.TrimSpace(name)
	for _, h := range p.HTMX {
		if !strings.Contains(h.Path, "/category/") || !strings.Contains(h.Path, "/switch") {
			continue
		}
		if strings.Contains(h.Label, want) {
			return h.Path
		}
	}
	return ""
}

// IsBlockedAgentPath reports routes test agents must not use.
func IsBlockedAgentPath(path string) bool {
	return strings.HasPrefix(path, "/admin/")
}

// IsBlockedAgentForm reports POST targets test agents must not submit.
func IsBlockedAgentForm(path string) bool {
	switch path {
	case "/auth/user/settings/password":
		return true
	default:
		return false
	}
}

// FilterSensitiveForms removes forms and HTMX targets agents should not use.
func FilterSensitiveForms(p PageAffordances) PageAffordances {
	out := p
	out.Forms = nil
	for _, f := range p.Forms {
		if IsBlockedAgentForm(f.Action) {
			continue
		}
		out.Forms = append(out.Forms, f)
	}
	out.HTMX = nil
	for _, h := range p.HTMX {
		if IsBlockedAgentForm(h.Path) {
			continue
		}
		out.HTMX = append(out.HTMX, h)
	}
	return out
}

// FilterHTMXPrefix removes in-page controls whose path starts with prefix.
func FilterHTMXPrefix(p PageAffordances, prefix string) PageAffordances {
	out := p
	out.HTMX = nil
	for _, h := range p.HTMX {
		if strings.HasPrefix(h.Path, prefix) {
			continue
		}
		out.HTMX = append(out.HTMX, h)
	}
	return out
}

// IsModalPath reports HTMX endpoints that open, close, or save modals.
func IsModalPath(path string) bool {
	if strings.HasPrefix(path, "/api/modal-remove/") {
		return true
	}
	if strings.Contains(path, "-modal") {
		return true
	}
	if strings.Contains(path, "/modal") {
		return true
	}
	return false
}

// IsStuckPage reports when the composed DOM has almost no navigation options.
func IsStuckPage(p PageAffordances) bool {
	if len(p.AdCards) > 0 || len(p.Links) > 0 {
		return false
	}
	if len(p.Forms) > 0 {
		return false
	}
	if len(p.HTMX) > 0 && !onlyModalHTMX(p.HTMX) {
		return false
	}
	return true
}

func onlyModalHTMX(actions []HTMXAction) bool {
	if len(actions) == 0 {
		return false
	}
	for _, h := range actions {
		if !IsModalPath(h.Path) {
			return false
		}
	}
	return true
}

// IsAuthEntryPath reports login/register pages that log out or re-register.
func IsAuthEntryPath(path string) bool {
	return path == "/login" || path == "/register"
}

// FilterAffordancesForPlanner returns a copy omitting modal HTMX when breaking loops.
func FilterAffordancesForPlanner(p PageAffordances, omitModals bool) PageAffordances {
	if !omitModals {
		return p
	}
	out := p
	out.HTMX = nil
	for _, h := range p.HTMX {
		if !IsModalPath(h.Path) {
			out.HTMX = append(out.HTMX, h)
		}
	}
	return out
}

// FilterForLoggedInUser removes login/register affordances for authenticated agents.
func FilterForLoggedInUser(p PageAffordances) PageAffordances {
	out := p
	out.Links = nil
	for _, l := range p.Links {
		if IsAuthEntryPath(l.Href) || IsBlockedAgentPath(l.Href) {
			continue
		}
		out.Links = append(out.Links, l)
	}
	out.Forms = nil
	for _, f := range p.Forms {
		if strings.HasPrefix(f.Action, "/api/register/") ||
			f.Action == "/api/login" {
			continue
		}
		out.Forms = append(out.Forms, f)
	}
	return out
}

// Known auth-area pages reachable when logged in (not always linked on home).
var knownAuthPaths = []string{
	"/auth/ad/new",
	"/auth/user/myads",
	"/auth/user/messages",
	"/auth/user/settings",
	"/auth/welcome",
}

// IsAuthNavPath reports authenticated app pages.
func IsAuthNavPath(path string) bool {
	return strings.HasPrefix(path, "/auth/")
}

// AppendKnownAuthPaths adds common auth routes for planner/validation.
func AppendKnownAuthPaths(p PageAffordances) PageAffordances {
	seen := p.AllowedPaths()
	out := p
	for _, path := range knownAuthPaths {
		if seen[path] {
			continue
		}
		out.Links = append(out.Links, Link{Href: path, Label: path})
	}
	return out
}

// PageLooksLoggedOut reports nav hints that the viewer is logged out.
func PageLooksLoggedOut(p PageAffordances) bool {
	for _, l := range p.Links {
		if l.Href == "/login" || l.Href == "/register" {
			return true
		}
	}
	return false
}

// PageLooksLoggedIn reports nav hints that the viewer is logged in.
func PageLooksLoggedIn(p PageAffordances) bool {
	if p.LoggedIn {
		return true
	}
	for _, l := range p.Links {
		if l.Href == "/logout" || strings.HasPrefix(l.Href, "/auth/user/") {
			return true
		}
	}
	for _, h := range p.HTMX {
		if h.Path == "/auth/user/menu" || strings.HasPrefix(h.Path, "/auth/sse") {
			return true
		}
	}
	return false
}

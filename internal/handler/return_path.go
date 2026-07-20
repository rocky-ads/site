package handler

import "strings"

// safeReturnPath returns raw if it is a same-origin relative path safe to
// redirect to after login. Open redirects and API paths are rejected.
func safeReturnPath(raw string) string {
	if raw == "" || raw[0] != '/' || (len(raw) > 1 && raw[1] == '/') {
		return ""
	}
	path := raw
	if i := strings.IndexAny(raw, "?#"); i >= 0 {
		path = raw[:i]
	}
	switch {
	case path == "/login", strings.HasPrefix(path, "/login/"):
		return ""
	case path == "/register", strings.HasPrefix(path, "/register/"):
		return ""
	case strings.HasPrefix(path, "/api/"):
		return ""
	}
	return raw
}

func loginRedirectPath(raw string) string {
	if p := safeReturnPath(raw); p != "" {
		return p
	}
	return "/"
}

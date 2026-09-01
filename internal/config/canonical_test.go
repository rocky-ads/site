package config

import "testing"

func TestCanonicalURL(t *testing.T) {
	prev := PublicSiteURL
	PublicSiteURL = "https://rockyads.com"
	t.Cleanup(func() { PublicSiteURL = prev })

	tests := []struct {
		path, want string
	}{
		{"", "https://rockyads.com/"},
		{"/", "https://rockyads.com/"},
		{"/login", "https://rockyads.com/login"},
		{"/login?return=/", "https://rockyads.com/login"},
		{"about", "https://rockyads.com/about"},
	}
	for _, tt := range tests {
		if got := CanonicalURL(tt.path); got != tt.want {
			t.Errorf("CanonicalURL(%q) = %q, want %q",
				tt.path, got, tt.want)
		}
	}
}

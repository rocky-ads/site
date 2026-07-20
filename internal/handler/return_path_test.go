package handler

import "testing"

func TestSafeReturnPath(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"/ad/730", "/ad/730"},
		{"/ad/730?x=1", "/ad/730?x=1"},
		{"//evil.com", ""},
		{"https://evil.com", ""},
		{"/login", ""},
		{"/login?return=/", ""},
		{"/register", ""},
		{"/register/verify", ""},
		{"/api/login", ""},
		{"/auth/user/messages", "/auth/user/messages"},
		{"/", "/"},
	}
	for _, tt := range tests {
		if got := safeReturnPath(tt.in); got != tt.want {
			t.Errorf("safeReturnPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLoginRedirectPath(t *testing.T) {
	if got := loginRedirectPath("/ad/1"); got != "/ad/1" {
		t.Fatalf("got %q", got)
	}
	if got := loginRedirectPath("//evil"); got != "/" {
		t.Fatalf("got %q", got)
	}
	if got := loginRedirectPath(""); got != "/" {
		t.Fatalf("got %q", got)
	}
}

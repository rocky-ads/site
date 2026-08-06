package config

import "testing"

func TestTrimURLBase(t *testing.T) {
	if got := trimURLBase(" https://rockyads.com/ "); got != "https://rockyads.com" {
		t.Fatalf("got %q", got)
	}
}

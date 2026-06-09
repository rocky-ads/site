package local

import "testing"

func TestIsLoggedIn(t *testing.T) {
	if IsLoggedIn(0) {
		t.Fatal("user id 0 should not be logged in")
	}
	if !IsLoggedIn(1) {
		t.Fatal("non-zero user id should be logged in")
	}
}

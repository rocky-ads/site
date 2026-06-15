package user

import "testing"

func TestIsTestPhoneE64(t *testing.T) {
	tests := []struct {
		phone string
		want  bool
	}{
		{"+15550100123", true},
		{"+15550109999", true},
		{"+1555010012", false},
		{"+155501001234", false},
		{"+15550110123", false},
		{"+14155551234", false},
	}
	for _, tt := range tests {
		if got := IsTestPhoneE64(tt.phone); got != tt.want {
			t.Errorf("IsTestPhoneE64(%q) = %v, want %v", tt.phone, got, tt.want)
		}
	}
}

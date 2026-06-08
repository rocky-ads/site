package phoneformat

import "testing"

func TestDisplay(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"+15035238780", "(503) 523-8780"},
		{"", ""},
		{"not-a-phone", "not-a-phone"},
	}
	for _, tc := range tests {
		got := Display(tc.in)
		if got != tc.want {
			t.Errorf("Display(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

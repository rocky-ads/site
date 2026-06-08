package ad

import "testing"

func TestSanitizeAdText(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"plain ascii", "plain ascii"},
		{"caf\u00e9", "caf\u00e9"},
		{"line one\nline two", "line one\nline two"},
		{"smart \u2018quote\u2019", "smart 'quote'"},
		{"em\u2014dash", "em-dash"},
		{"3.2\u00a0L", "3.2 L"},
		{"trailing junk\uFFFC", "trailing junk"},
		{"\u200Bhidden", "hidden"},
		{"tab\there", "tab here"},
		{"spoof\u001ehistory", "spoofhistory"},
		{"spoof\u001fhistory", "spoofhistory"},
	}
	for _, tt := range tests {
		if got := SanitizeAdText(tt.in); got != tt.want {
			t.Errorf("SanitizeAdText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

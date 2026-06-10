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

func TestTitleContainsEmoji(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"2020 Honda Civic", false},
		{"caf\u00e9 sale", false},
		{"Huge sale \U0001F95A", true},
		{"Party \U0001F389", true},
		{"Warning \u26A0", true},
		{"US flag \U0001F1FA\U0001F1F8", true},
	}
	for _, tt := range tests {
		if got := TitleContainsEmoji(tt.in); got != tt.want {
			t.Errorf("TitleContainsEmoji(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

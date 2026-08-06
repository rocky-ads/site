package handler

import "testing"

func TestSMSOptKeyword(t *testing.T) {
	tests := []struct {
		body     string
		optedOut bool
		ok       bool
	}{
		{"STOP", true, true},
		{" stop ", true, true},
		{"STOPALL", true, true},
		{"unsubscribe", true, true},
		{"CANCEL", true, true},
		{"END", true, true},
		{"QUIT", true, true},
		{"REVOKE", true, true},
		{"OPTOUT", true, true},
		{"START", false, true},
		{"yes", false, true},
		{"UNSTOP", false, true},
		{"RECOVER ABC123", false, false},
		{"hello", false, false},
		{"", false, false},
		{"STOP PLEASE", false, false},
	}
	for _, tt := range tests {
		optedOut, ok := smsOptKeyword(tt.body)
		if ok != tt.ok || optedOut != tt.optedOut {
			t.Fatalf("%q: got optedOut=%v ok=%v, want optedOut=%v ok=%v",
				tt.body, optedOut, ok, tt.optedOut, tt.ok)
		}
	}
}

func TestApplySMSPreferenceNonKeyword(t *testing.T) {
	if applySMSPreferenceFromInbound("+15551234567", "hello") {
		t.Fatal("non-keyword should not be handled")
	}
}

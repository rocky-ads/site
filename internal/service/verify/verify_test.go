package verify

import (
	"errors"
	"strings"
	"testing"
)

func TestMapTwilioErr(t *testing.T) {
	tests := []struct {
		msg  string
		want error
	}{
		{"Status: 404, Code: 60200", ErrInvalidCode},
		{"was not found", ErrInvalidCode},
		{"21610 blacklisted", ErrBlocked},
		{"number opted out", ErrBlocked},
		{"geo permissions", ErrRejected},
		{"fraud guard", ErrRejected},
		{"unknown boom", ErrRejected},
	}
	for _, tt := range tests {
		got := mapTwilioErr(errors.New(tt.msg))
		if !errors.Is(got, tt.want) {
			t.Fatalf("msg=%q got=%v want=%v", tt.msg, got, tt.want)
		}
		if !strings.Contains(got.Error(), tt.msg) &&
			tt.want != ErrInvalidCode {
			// ErrInvalidCode may return bare sentinel
			if got != ErrInvalidCode {
				t.Fatalf("msg=%q missing wrap: %v", tt.msg, got)
			}
		}
	}
}

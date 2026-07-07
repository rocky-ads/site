package password

import "testing"

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name    string
		pass    string
		wantErr bool
	}{
		{
			name: "google style alphanumeric",
			pass: "Kj8mN2pQxRvLwYz",
		},
		{
			name: "google style with symbols",
			pass: "aBc9$xK!mN2pQ7wR",
		},
		{
			name: "long password manager passphrase",
			pass: "correct-horse-battery-staple-2024",
		},
		{
			name:    "too short",
			pass:    "short1",
			wantErr: true,
		},
		{
			name: "repeated character",
			pass: "aaaaaaaa",
		},
		{
			name:    "over max length",
			pass:    "abcdefghijklmnop" + string(make([]byte, MaxPasswordLength)),
			wantErr: true,
		},
		{
			name: "at minimum length",
			pass: "abcdefgh",
		},
		{
			name: "at maximum length",
			pass: "abcdefghijklmnop" +
				"qrstuvwxyzABCDEF" +
				"GHIJKLMNOPQRSTUV" +
				"WXYZ0123456789ab" +
				"cdefghijklmnopqr" +
				"stuvwxyzABCDEFGH" +
				"IJKLMNOPQRSTUVWX" +
				"YZ0123456789abcd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.pass)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePasswordStrength() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}

package password

import (
	"errors"
	"fmt"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

var (
	strengthTooShort = fmt.Sprintf(
		"Password must be at least %d characters.", MinPasswordLength)
	strengthTooLong = fmt.Sprintf(
		"Password must be at most %d characters.", MaxPasswordLength)
)

// StrengthRequirements is the full password rule text for user-facing errors.
var StrengthRequirements = fmt.Sprintf(
	"Password must be between %d and %d characters.",
	MinPasswordLength, MaxPasswordLength)

func ValidatePasswordConfirmation(password, confirmation string) error {
	if password != confirmation {
		return fmt.Errorf("passwords do not match")
	}
	return nil
}

func ValidatePasswordStrength(password string) error {
	n := len(password)
	if n < MinPasswordLength {
		return errors.New(strengthTooShort)
	}
	if n > MaxPasswordLength {
		return errors.New(strengthTooLong)
	}
	return nil
}

// ValidatePasswordChange validates a password change operation
func ValidatePasswordChange(currentPassword, newPassword,
	confirmPassword string) error {
	// Check if current password is provided
	if currentPassword == "" {
		return fmt.Errorf("current password is required")
	}

	// Check if new password is provided
	if newPassword == "" {
		return fmt.Errorf("new password is required")
	}

	// Check if new password is different from current
	if currentPassword == newPassword {
		return fmt.Errorf("new password must be different from current password")
	}

	// Validate password confirmation
	if err := ValidatePasswordConfirmation(newPassword, confirmPassword); err != nil {
		return err
	}

	// Validate new password strength
	if err := ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	return nil
}

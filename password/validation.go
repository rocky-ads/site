package password

import (
	"errors"
	"fmt"
	"unicode"
)

// StrengthRequirements is the full password rule text for user-facing errors.
const StrengthRequirements = "Password must be at least 8 characters and include " +
	"uppercase and lowercase letters, a number, and a special character (e.g., !@#$%^&*)."

func ValidatePasswordConfirmation(password, confirmation string) error {
	if password != confirmation {
		return fmt.Errorf("passwords do not match")
	}
	return nil
}

func ValidatePasswordStrength(password string) error {
	okLen := len(password) >= 8
	hasNumber := false
	hasUpper := false
	hasLower := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			if unicode.IsUpper(char) {
				hasUpper = true
			}
			if unicode.IsLower(char) {
				hasLower = true
			}
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) &&
			!unicode.IsSpace(char) && char >= 33 && char <= 126 {
			hasSpecial = true
		}
	}

	if okLen && hasUpper && hasLower && hasNumber && hasSpecial {
		return nil
	}

	return errors.New(StrengthRequirements)
}

// ValidatePasswordChange validates a password change operation
func ValidatePasswordChange(currentPassword, newPassword, confirmPassword string) error {
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

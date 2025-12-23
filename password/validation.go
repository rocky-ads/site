package password

import (
	"fmt"
	"unicode"
)

func ValidatePasswordConfirmation(password, confirmation string) error {
	if password != confirmation {
		return fmt.Errorf("passwords do not match")
	}
	return nil
}

// ValidatePasswordStrength checks if a password meets minimum requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Check for at least one letter, one number, one uppercase, one lowercase, and one special character
	hasLetter := false
	hasNumber := false
	hasUpper := false
	hasLower := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
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
		// Check for special characters (printable ASCII that's not letter, number, or space)
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) && !unicode.IsSpace(char) && char >= 33 && char <= 126 {
			hasSpecial = true
		}
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character (e.g., !@#$%%^&*)")
	}

	return nil
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

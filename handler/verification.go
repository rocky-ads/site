package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"time"

	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/logger"
)

const (
	// CodeLength is the length of verification codes
	CodeLength = 6
	// CodeExpiry is how long codes are valid
	CodeExpiry = 10 * time.Minute
	// MaxAttempts is the maximum verification attempts allowed
	MaxAttempts = 3
	// MaxFailedVerifications is the maximum failed verification attempts before account cleanup
	MaxFailedVerifications = 5
	// VerificationWindow is the time window for tracking failed verifications
	VerificationWindow = 24 * time.Hour
)

type phoneVerification struct {
	ID               int       `json:"id"`
	Phone            string    `json:"phone"`
	VerificationCode string    `json:"verification_code"`
	Attempts         int       `json:"attempts"`
	CreatedAt        time.Time `json:"created_at"`
}

func generateVerificationCode() (string, error) {
	code := ""
	for i := 0; i < CodeLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		code += fmt.Sprintf("%d", num.Int64())
	}
	return code, nil
}

func storeVerificationCode(phone, code string) error {
	_, err := db.Exec(`
		INSERT INTO phone_verification (phone, verification_code, attempts) 
		VALUES ($1, $2, 0)
	`, phone, code)

	if err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}

	logger.Info("Created verification code",
		"component", "verification", "phone", phone)

	// Lazy cleanup: remove expired codes when storing a new one
	// Ignore cleanup errors to avoid blocking code storage
	if err := cleanupExpiredCodes(); err != nil {
		logger.Warn("Failed to cleanup expired codes during code storage",
			"error", err, "component", "verification")
	}

	return nil
}

func getPhoneVerification(phone string) (phoneVerification, error) {
	var pv phoneVerification
	// Calculate expiry threshold: records created before this time are expired
	expiryThreshold := time.Now().UTC().Add(-CodeExpiry)
	err := db.QueryJSON(&pv, `
		SELECT json_object(
			'id', id,
			'phone', phone,
			'verification_code', verification_code,
			'attempts', attempts,
			'created_at', created_at
		)
		FROM phone_verification 
		WHERE phone = $1 
			AND created_at > $2
			AND attempts < $3
		ORDER BY created_at DESC 
		LIMIT 1
	`, phone, expiryThreshold, MaxAttempts)
	if err != nil {
		return phoneVerification{}, fmt.Errorf("verification code not found: %w", err)
	}

	return pv, nil
}

func validateVerificationCode(phone, code string) (bool, error) {
	vc, err := getPhoneVerification(phone)
	if err != nil {
		// If no valid record found, it could be expired or max attempts exceeded
		// Track this for potential account cleanup
		cleanupStaleVerifications(phone)
		return false, err
	}

	// Increment attempts
	_, err = db.Exec(`
		UPDATE phone_verification 
		SET attempts = attempts + 1 
		WHERE id = $1
	`, vc.ID)
	if err != nil {
		logger.Error("Failed to update attempts",
			"error", err, "component", "verification")
		return false, err
	}

	codeBytes := []byte(vc.VerificationCode)
	inputBytes := []byte(code)

	// Ensure both slices are the same length for constant-time comparison
	if len(codeBytes) != len(inputBytes) {
		logger.Warn("Invalid code (length mismatch)",
			"component", "verification", "phone", phone)
		return false, nil
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(codeBytes, inputBytes) == 1 {
		logger.Info("Code validated successfully",
			"component", "verification", "phone", phone)
		return true, nil
	}

	logger.Warn("Invalid code", "component", "verification", "phone", phone)
	return false, nil
}

func cleanupExpiredCodes() error {
	// Delete codes where created_at + CodeExpiry < now
	expiryThreshold := time.Now().UTC().Add(-CodeExpiry)
	_, err := db.Exec(`
		DELETE FROM phone_verification 
		WHERE created_at < $1
	`, expiryThreshold)

	if err != nil {
		return fmt.Errorf("failed to cleanup expired codes: %w", err)
	}

	logger.Info("Cleaned up expired verification codes",
		"component", "verification")
	return nil
}

func markPhoneVerified(userID int) error {
	_, err := db.Exec(`
		UPDATE users 
		SET phone_verified = 1 
		WHERE id = $1
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to mark phone verified: %w", err)
	}

	logger.Info("Marked phone verified for user",
		"component", "verification", "userID", userID)
	return nil
}

// invalidateVerificationCodes invalidates all verification codes for a phone number
// This is called when a user replies STOP or when SMS delivery fails
func invalidateVerificationCodes(phone string) error {
	_, err := db.Exec(`
		DELETE FROM phone_verification 
		WHERE phone = $1
	`, phone)

	if err != nil {
		return fmt.Errorf("failed to invalidate verification codes for %s: %w", phone, err)
	}

	logger.Info("Invalidated all verification codes for phone",
		"component", "verification", "phone", phone)
	return nil
}

func cleanupStaleVerifications(phone string) error {
	// Count failed verifications in the last 24 hours
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM phone_verification 
		WHERE phone = $1 AND created_at > $2
	`, phone, time.Now().UTC().Add(-VerificationWindow)).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to count failed verifications for %s: %w", phone, err)
	}

	// If we've exceeded the threshold, clean up the account
	if count >= MaxFailedVerifications {
		logger.Warn("Max failed verifications exceeded, cleaning up account",
			"component", "verification", "phone", phone)
		return cleanupFailedAccount(phone)
	}

	return nil
}

func cleanupFailedAccount(phone string) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete verification codes
	_, err = tx.Exec(`DELETE FROM phone_verification WHERE phone = $1`, phone)
	if err != nil {
		return fmt.Errorf("failed to delete verification codes: %w", err)
	}

	// Delete any partial user records (users created but not verified)
	// Use phone_hash since users table stores phone_hash, not plain phone
	phoneHash := db.HashString(phone)
	_, err = tx.Exec(`DELETE FROM users WHERE phone_hash = $1 AND phone_verified = 0`, phoneHash)
	if err != nil {
		return fmt.Errorf("failed to delete unverified user: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	logger.Info("Successfully cleaned up failed account",
		"component", "verification", "phone", phone)
	return nil
}

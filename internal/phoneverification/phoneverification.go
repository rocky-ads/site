package phoneverification

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
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

	PurposeRegister    = "register"
	PurposeChangePhone = "change_phone"
)

type phoneVerification struct {
	ID               int
	PhoneE64         string
	VerificationCode string
	Purpose          string
	UserID           sql.NullInt64
	Attempts         int
	CreatedAt        time.Time
}

func GenerateCode() (string, error) {
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

// StoreCode stores a verification code bound to a purpose.
// For PurposeChangePhone, userID must be non-nil.
func StoreCode(phoneE64, code, purpose string, userID *int) error {
	if err := validatePurpose(purpose, userID); err != nil {
		return err
	}

	var uid any
	if userID != nil {
		uid = *userID
	}

	_, err := db.Exec(`
		INSERT INTO phone_verification (
			phone_e64, verification_code, purpose, user_id, attempts
		) VALUES ($1, $2, $3, $4, 0)
	`, phoneE64, code, purpose, uid)

	if err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}

	logger.Info("Created verification code",
		"component", "verification", "phoneE64", phoneE64,
		"purpose", purpose)

	if err := cleanupExpiredCodes(); err != nil {
		logger.Warn("Failed to cleanup expired codes during code storage",
			"error", err, "component", "verification")
	}

	return nil
}

func validatePurpose(purpose string, userID *int) error {
	switch purpose {
	case PurposeRegister:
		if userID != nil {
			return fmt.Errorf("register verification must not include user_id")
		}
	case PurposeChangePhone:
		if userID == nil {
			return fmt.Errorf("change_phone verification requires user_id")
		}
	default:
		return fmt.Errorf("unknown verification purpose: %s", purpose)
	}
	return nil
}

func getPhoneVerification(phoneE64, purpose string,
	userID *int) (phoneVerification, error) {
	if err := validatePurpose(purpose, userID); err != nil {
		return phoneVerification{}, err
	}

	var pv phoneVerification
	expiryThreshold := time.Now().UTC().Add(-CodeExpiry)

	var err error
	if purpose == PurposeChangePhone {
		err = db.QueryRow(`
			SELECT id, phone_e64, verification_code, purpose, user_id,
				attempts, created_at
			FROM phone_verification
			WHERE phone_e64 = $1
				AND purpose = $2
				AND user_id = $3
				AND created_at > $4
				AND attempts < $5
			ORDER BY created_at DESC
			LIMIT 1
		`, phoneE64, purpose, *userID, expiryThreshold, MaxAttempts).Scan(
			&pv.ID,
			&pv.PhoneE64,
			&pv.VerificationCode,
			&pv.Purpose,
			&pv.UserID,
			&pv.Attempts,
			&pv.CreatedAt,
		)
	} else {
		err = db.QueryRow(`
			SELECT id, phone_e64, verification_code, purpose, user_id,
				attempts, created_at
			FROM phone_verification
			WHERE phone_e64 = $1
				AND purpose = $2
				AND user_id IS NULL
				AND created_at > $3
				AND attempts < $4
			ORDER BY created_at DESC
			LIMIT 1
		`, phoneE64, purpose, expiryThreshold, MaxAttempts).Scan(
			&pv.ID,
			&pv.PhoneE64,
			&pv.VerificationCode,
			&pv.Purpose,
			&pv.UserID,
			&pv.Attempts,
			&pv.CreatedAt,
		)
	}
	if err != nil {
		return phoneVerification{}, fmt.Errorf("verification code not found: %w", err)
	}

	return pv, nil
}

func codesMatch(stored, input string) bool {
	storedBytes := []byte(stored)
	inputBytes := []byte(input)
	if len(storedBytes) != len(inputBytes) {
		return false
	}
	return subtle.ConstantTimeCompare(storedBytes, inputBytes) == 1
}

func bumpAttempts(id int) error {
	_, err := db.Exec(`
		UPDATE phone_verification
		SET attempts = attempts + 1
		WHERE id = $1
	`, id)
	if err != nil {
		logger.Error("Failed to update attempts",
			"error", err, "component", "verification")
		return err
	}
	return nil
}

func ValidateCode(phoneE64, code, purpose string, userID *int) (bool, error) {
	vc, err := getPhoneVerification(phoneE64, purpose, userID)
	if err != nil {
		cleanupStaleVerifications(phoneE64)
		return false, err
	}

	if codesMatch(vc.VerificationCode, code) {
		logger.Info("Code validated successfully",
			"component", "verification", "phoneE64", phoneE64,
			"purpose", purpose)
		return true, nil
	}

	if err := bumpAttempts(vc.ID); err != nil {
		return false, err
	}

	logger.Warn("Invalid code", "component", "verification", "phoneE64", phoneE64)
	return false, nil
}

// ConsumeCode validates a code and deletes it so it cannot be reused.
// Concurrent consumers: only the first successful delete wins.
func ConsumeCode(phoneE64, code, purpose string, userID *int) (bool, error) {
	vc, err := getPhoneVerification(phoneE64, purpose, userID)
	if err != nil {
		cleanupStaleVerifications(phoneE64)
		return false, err
	}

	if !codesMatch(vc.VerificationCode, code) {
		if err := bumpAttempts(vc.ID); err != nil {
			return false, err
		}
		logger.Warn("Invalid code", "component", "verification",
			"phoneE64", phoneE64)
		return false, nil
	}

	res, err := db.Exec(`
		DELETE FROM phone_verification WHERE id = $1
	`, vc.ID)
	if err != nil {
		return false, fmt.Errorf("failed to consume verification code: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check consumed code: %w", err)
	}
	if n == 0 {
		return false, nil
	}

	logger.Info("Code consumed successfully",
		"component", "verification", "phoneE64", phoneE64,
		"purpose", purpose)
	return true, nil
}

func cleanupExpiredCodes() error {
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

// InvalidateCodes invalidates all verification codes for a phone number.
// Used when a user replies STOP or when SMS delivery fails.
func InvalidateCodes(phoneE64 string) error {
	if err := invalidateAllCodes(phoneE64, db.Exec); err != nil {
		return err
	}

	logger.Info("Invalidated all verification codes for phone",
		"component", "verification", "phoneE64", phoneE64)
	return nil
}

// InvalidateCodesTx invalidates verification codes for a phone and purpose
// within an existing transaction.
func InvalidateCodesTx(tx *sql.Tx, phoneE64, purpose string,
	userID *int) error {
	return invalidateCodesForPurpose(phoneE64, purpose, userID, tx.Exec)
}

// InvalidateCodesForPurpose invalidates codes for a phone and purpose.
func InvalidateCodesForPurpose(phoneE64, purpose string, userID *int) error {
	if err := invalidateCodesForPurpose(phoneE64, purpose, userID, db.Exec); err != nil {
		return err
	}
	logger.Info("Invalidated verification codes for purpose",
		"component", "verification", "phoneE64", phoneE64,
		"purpose", purpose)
	return nil
}

func invalidateAllCodes(phoneE64 string,
	exec func(string, ...any) (sql.Result, error)) error {
	_, err := exec(`
		DELETE FROM phone_verification
		WHERE phone_e64 = $1
	`, phoneE64)
	if err != nil {
		return fmt.Errorf("failed to invalidate verification codes for %s: %w",
			phoneE64, err)
	}
	return nil
}

func invalidateCodesForPurpose(phoneE64, purpose string, userID *int,
	exec func(string, ...any) (sql.Result, error)) error {
	if err := validatePurpose(purpose, userID); err != nil {
		return err
	}

	var err error
	if purpose == PurposeChangePhone {
		_, err = exec(`
			DELETE FROM phone_verification
			WHERE phone_e64 = $1 AND purpose = $2 AND user_id = $3
		`, phoneE64, purpose, *userID)
	} else {
		_, err = exec(`
			DELETE FROM phone_verification
			WHERE phone_e64 = $1 AND purpose = $2 AND user_id IS NULL
		`, phoneE64, purpose)
	}
	if err != nil {
		return fmt.Errorf("failed to invalidate verification codes for %s: %w",
			phoneE64, err)
	}
	return nil
}

func cleanupStaleVerifications(phoneE64 string) error {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM phone_verification
		WHERE phone_e64 = $1 AND created_at > $2
	`, phoneE64, time.Now().UTC().Add(-VerificationWindow)).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to count failed verifications for %s: %w",
			phoneE64, err)
	}

	if count >= MaxFailedVerifications {
		logger.Warn("Max failed verifications exceeded, cleaning up account",
			"component", "verification", "phoneE64", phoneE64)
		return cleanupFailedAccount(phoneE64)
	}

	return nil
}

func cleanupFailedAccount(phoneE64 string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := invalidateAllCodes(phoneE64, tx.Exec); err != nil {
		return fmt.Errorf("failed to delete verification codes: %w", err)
	}

	phoneHash := db.HashString(phoneE64)
	_, err = tx.Exec(
		`DELETE FROM users WHERE phone_hash = $1 AND phone_verified = 0`,
		phoneHash,
	)
	if err != nil {
		return fmt.Errorf("failed to delete unverified user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	logger.Info("Successfully cleaned up failed account",
		"component", "verification", "phone", phoneE64)
	return nil
}

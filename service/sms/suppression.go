package sms

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/message"
)

// ShouldSuppressSMS checks if SMS should be suppressed for a recipient
// Returns true if:
// - Last SMS was sent within the suppression window (10 minutes)
// - User has read all conversations (no unread messages)
func ShouldSuppressSMS(recipientUserID int) (bool, error) {
	// Check if last SMS was sent within suppression window
	lastSMSSent, err := getLastSMSSent(recipientUserID)
	if err != nil {
		return false, fmt.Errorf("failed to get last SMS sent time: %w", err)
	}

	if lastSMSSent != nil {
		windowDuration := time.Duration(config.SMSSuppressionWindowMinutes) * time.Minute
		if time.Since(*lastSMSSent) < windowDuration {
			return true, nil
		}
	}

	// Check if user has read all conversations (no unread messages)
	hasUnread, err := message.GetHasUnread(recipientUserID)
	if err != nil {
		return false, fmt.Errorf("failed to check unread status: %w", err)
	}

	if !hasUnread {
		return true, nil
	}

	return false, nil
}

// GetUnreadMessageCount returns the count of unread messages across all conversations
func GetUnreadMessageCount(userID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM conversations
		WHERE (owner_id = ? AND owner_has_unread = 1)
		   OR (enquirer_id = ? AND enquirer_has_unread = 1)
	`
	var count int
	err := db.QueryRow(query, userID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread message count: %w", err)
	}
	return count, nil
}

// UpdateLastSMSSent updates the last_sms_sent_at timestamp for a user
func UpdateLastSMSSent(userID int) error {
	_, err := db.Exec(`
		UPDATE users
		SET last_sms_sent_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to update last SMS sent timestamp: %w", err)
	}
	return nil
}

// getLastSMSSent retrieves the last_sms_sent_at timestamp for a user
func getLastSMSSent(userID int) (*time.Time, error) {
	var lastSMSSent sql.NullTime
	err := db.QueryRow(`
		SELECT last_sms_sent_at
		FROM users
		WHERE id = ?
	`, userID).Scan(&lastSMSSent)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last SMS sent time: %w", err)
	}

	if !lastSMSSent.Valid {
		return nil, nil
	}

	return &lastSMSSent.Time, nil
}

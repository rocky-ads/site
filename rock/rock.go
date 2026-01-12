package rock

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rocky-ads/site/db"
)

var ErrRockNotFound = errors.New("rock not found")
var ErrMaxRocksReached = errors.New("user has reached maximum outstanding rocks")
var ErrRockAlreadyThrown = errors.New("a rock has already been thrown at this conversation")

// ThrowRock throws a rock at a conversation, making it public
// If enquirer throws: rock_thrower_id = enquirer_id (bound to ad)
// If owner throws: rock_thrower_id = owner_id (bound to enquirer)
func ThrowRock(userID, conversationID int) error {
	// Get conversation to determine owner and enquirer
	conv, err := getConversationForRock(conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Verify user is participant
	if conv.OwnerID != userID && conv.EnquirerID != userID {
		return fmt.Errorf("only conversation participants can throw rocks")
	}

	// Check if user already has 3 outstanding rocks
	count, err := GetUserRockCount(userID)
	if err != nil {
		return fmt.Errorf("failed to check rock count: %w", err)
	}
	if count >= 3 {
		return ErrMaxRocksReached
	}

	// Check if ANY rock already exists on this conversation (only one rock per conversation)
	rockCount, err := GetRockCountForConversation(conversationID)
	if err != nil {
		return fmt.Errorf("failed to check existing rocks: %w", err)
	}
	if rockCount > 0 {
		return ErrRockAlreadyThrown
	}

	// Update conversation with rock info (making it public)
	_, err = db.Exec(`
		UPDATE conversations
		SET rock_thrower_id = ?, rock_thrown_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to throw rock: %w", err)
	}

	return nil
}

// getConversationForRock gets conversation details needed for rock throwing
func getConversationForRock(conversationID int) (struct {
	OwnerID    int
	EnquirerID int
}, error) {
	var conv struct {
		OwnerID    int
		EnquirerID int
	}
	err := db.QueryRow(`
		SELECT owner_id, enquirer_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&conv.OwnerID, &conv.EnquirerID)
	if err != nil {
		return conv, fmt.Errorf("failed to get conversation: %w", err)
	}
	return conv, nil
}

// UnthrowRock removes a rock from a conversation
func UnthrowRock(userID, conversationID int) error {
	// Verify user is participant and threw the rock
	var rockThrowerID sql.NullInt64
	err := db.QueryRow(`
		SELECT rock_thrower_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&rockThrowerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRockNotFound
		}
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	if !rockThrowerID.Valid {
		return ErrRockNotFound
	}

	if int(rockThrowerID.Int64) != userID {
		return fmt.Errorf("only the rock thrower can remove the rock")
	}

	// Remove rock (making conversation private)
	_, err = db.Exec(`
		UPDATE conversations
		SET rock_thrower_id = NULL, rock_thrown_at = NULL
		WHERE id = ?
	`, conversationID)
	if err != nil {
		return fmt.Errorf("failed to remove rock: %w", err)
	}

	return nil
}

// GetPublicConversationsForAd returns public conversation IDs for an ad
// Only returns conversations with rocks bound to the ad (rock_thrower_id = enquirer_id)
func GetPublicConversationsForAd(adID int) ([]int, error) {
	query := `
		SELECT c.id
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.ad_id = ? AND c.rock_thrower_id IS NOT NULL AND c.rock_thrower_id = c.enquirer_id
		GROUP BY c.id
		ORDER BY COALESCE(MAX(m.created_at), c.rock_thrown_at) DESC
	`
	var conversationIDs []int
	err := db.Select(&conversationIDs, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public conversations for ad: %w", err)
	}
	return conversationIDs, nil
}

// GetPublicConversationIDByOrdinal returns the conversation ID at the given ordinal position (0-based)
// for public conversations with rocks bound to the ad (rock_thrower_id = enquirer_id), ordered by latest activity DESC
func GetPublicConversationIDByOrdinal(adID int, ordinal int) (int, error) {
	query := `
		SELECT c.id
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.ad_id = ? AND c.rock_thrower_id IS NOT NULL AND c.rock_thrower_id = c.enquirer_id
		GROUP BY c.id
		ORDER BY COALESCE(MAX(m.created_at), c.rock_thrown_at) DESC
		LIMIT 1 OFFSET ?
	`
	var conversationID int
	err := db.QueryRow(query, adID, ordinal).Scan(&conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("conversation not found at ordinal %d for ad %d", ordinal, adID)
		}
		return 0, fmt.Errorf("failed to get public conversation by ordinal: %w", err)
	}
	return conversationID, nil
}

// GetUserRockCount returns the count of outstanding rocks thrown by a user
func GetUserRockCount(userID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM conversations
		WHERE rock_thrower_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user rock count: %w", err)
	}
	return count, nil
}

// GetRockCountForConversation returns whether a rock exists for a conversation
func GetRockCountForConversation(conversationID int) (int, error) {
	var rockThrowerID sql.NullInt64
	err := db.QueryRow(`
		SELECT rock_thrower_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&rockThrowerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("conversation not found: %w", err)
		}
		return 0, fmt.Errorf("failed to get rock count for conversation: %w", err)
	}
	if !rockThrowerID.Valid {
		return 0, nil
	}
	return 1, nil
}

// HasUserThrownRock checks if a user has thrown a rock at a conversation
func HasUserThrownRock(userID, conversationID int) (bool, error) {
	var rockThrowerID sql.NullInt64
	err := db.QueryRow(`
		SELECT rock_thrower_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&rockThrowerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("conversation not found: %w", err)
		}
		return false, fmt.Errorf("failed to check if user threw rock: %w", err)
	}
	return rockThrowerID.Valid && int(rockThrowerID.Int64) == userID, nil
}

// GetConversationIDsForUserRocks returns conversation IDs for rocks bound to a user
// (where rock_thrower_id = owner_id, meaning owner threw rock at enquirer)
func GetConversationIDsForUserRocks(userID int) ([]int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE rock_thrower_id IS NOT NULL AND rock_thrower_id = owner_id AND enquirer_id = ?
		ORDER BY rock_thrown_at DESC
	`
	var conversationIDs []int
	err := db.Select(&conversationIDs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation IDs for user rocks: %w", err)
	}
	return conversationIDs, nil
}

// GetRockCountForUser returns the count of rocks bound to a user
// (rocks where owner threw at this user)
func GetRockCountForUser(userID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM conversations
		WHERE rock_thrower_id IS NOT NULL AND rock_thrower_id = owner_id AND enquirer_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get rock count for user: %w", err)
	}
	return count, nil
}

// GetConversationIDForUserRockByOrdinal returns the conversation ID at the given ordinal position (0-based)
// for rocks bound to a user (where rock_thrower_id = owner_id AND enquirer_id = userID), ordered by rock_thrown_at DESC
func GetConversationIDForUserRockByOrdinal(userID int, ordinal int) (int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE rock_thrower_id IS NOT NULL AND rock_thrower_id = owner_id AND enquirer_id = ?
		ORDER BY rock_thrown_at DESC
		LIMIT 1 OFFSET ?
	`
	var conversationID int
	err := db.QueryRow(query, userID, ordinal).Scan(&conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("conversation not found at ordinal %d for user %d", ordinal, userID)
		}
		return 0, fmt.Errorf("failed to get conversation ID for user rock by ordinal: %w", err)
	}
	return conversationID, nil
}

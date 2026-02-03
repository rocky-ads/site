package egg

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/rocky-ads/site/db"
)

var ErrEggNotFound = errors.New("egg not found")
var ErrMaxEggsReached = errors.New("user has reached maximum outstanding eggs")
var ErrEggAlreadyThrown = errors.New("an egg has already been thrown at this conversation")

// ThrowEgg throws an egg at a conversation, making it public
// If enquirer throws: egg_thrower_id = enquirer_id (bound to ad)
// If owner throws: egg_thrower_id = owner_id (bound to enquirer)
func ThrowEgg(userID, conversationID int) error {
	// Get conversation to determine owner and enquirer
	conv, err := getConversationForEgg(conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Verify user is participant
	if conv.OwnerID != userID && conv.EnquirerID != userID {
		return fmt.Errorf("only conversation participants can throw eggs")
	}

	// Check if user already has 3 outstanding eggs
	count, err := GetUserEggCount(userID)
	if err != nil {
		return fmt.Errorf("failed to check egg count: %w", err)
	}
	if count >= 3 {
		return ErrMaxEggsReached
	}

	// Check if ANY egg already exists on this conversation (only one egg per conversation)
	eggCount, err := GetEggCountForConversation(conversationID)
	if err != nil {
		return fmt.Errorf("failed to check existing eggs: %w", err)
	}
	if eggCount > 0 {
		return ErrEggAlreadyThrown
	}

	// Update conversation with egg info (making it public)
	_, err = db.Exec(`
		UPDATE conversations
		SET egg_thrower_id = ?, egg_thrown_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to throw egg: %w", err)
	}

	return nil
}

// getConversationForEgg gets conversation details needed for egg throwing
func getConversationForEgg(conversationID int) (struct {
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

// UnthrowEgg removes an egg from a conversation
func UnthrowEgg(userID, conversationID int) error {
	// Verify user is participant and threw the egg
	var eggThrowerID sql.NullInt64
	err := db.QueryRow(`
		SELECT egg_thrower_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&eggThrowerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrEggNotFound
		}
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	if !eggThrowerID.Valid {
		return ErrEggNotFound
	}

	if int(eggThrowerID.Int64) != userID {
		return fmt.Errorf("only the egg thrower can remove the egg")
	}

	// Remove egg (making conversation private)
	_, err = db.Exec(`
		UPDATE conversations
		SET egg_thrower_id = NULL, egg_thrown_at = NULL
		WHERE id = ?
	`, conversationID)
	if err != nil {
		return fmt.Errorf("failed to remove egg: %w", err)
	}

	return nil
}

// GetPublicConversationsForAd returns public conversation IDs for an ad
// Only returns conversations with eggs bound to the ad (egg_thrower_id = enquirer_id)
func GetPublicConversationsForAd(adID int) ([]int, error) {
	query := `
		SELECT c.id
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.ad_id = ? AND c.egg_thrower_id IS NOT NULL AND c.egg_thrower_id = c.enquirer_id
		GROUP BY c.id
		ORDER BY COALESCE(MAX(m.created_at), c.egg_thrown_at) DESC
	`
	var conversationIDs []int
	err := db.Select(&conversationIDs, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public conversations for ad: %w", err)
	}
	return conversationIDs, nil
}

// GetPublicConversationIDByOrdinal returns the conversation ID at the given ordinal position (0-based)
// for public conversations with eggs bound to the ad (egg_thrower_id = enquirer_id), ordered by latest activity DESC
func GetPublicConversationIDByOrdinal(adID int, ordinal int) (int, error) {
	query := `
		SELECT c.id
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.ad_id = ? AND c.egg_thrower_id IS NOT NULL AND c.egg_thrower_id = c.enquirer_id
		GROUP BY c.id
		ORDER BY COALESCE(MAX(m.created_at), c.egg_thrown_at) DESC
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

// GetUserEggCount returns the count of outstanding eggs thrown by a user
func GetUserEggCount(userID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM conversations
		WHERE egg_thrower_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user egg count: %w", err)
	}
	return count, nil
}

// GetEggCountForConversation returns whether an egg exists for a conversation
func GetEggCountForConversation(conversationID int) (int, error) {
	var eggThrowerID sql.NullInt64
	err := db.QueryRow(`
		SELECT egg_thrower_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&eggThrowerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("conversation not found: %w", err)
		}
		return 0, fmt.Errorf("failed to get egg count for conversation: %w", err)
	}
	if !eggThrowerID.Valid {
		return 0, nil
	}
	return 1, nil
}

// HasUserThrownEgg checks if a user has thrown an egg at a conversation
func HasUserThrownEgg(userID, conversationID int) (bool, error) {
	var eggThrowerID sql.NullInt64
	err := db.QueryRow(`
		SELECT egg_thrower_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&eggThrowerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("conversation not found: %w", err)
		}
		return false, fmt.Errorf("failed to check if user threw egg: %w", err)
	}
	return eggThrowerID.Valid && int(eggThrowerID.Int64) == userID, nil
}

// GetConversationIDsForUserEggs returns conversation IDs for eggs bound to a user
// (where egg_thrower_id = owner_id, meaning owner threw egg at enquirer)
func GetConversationIDsForUserEggs(userID int) ([]int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE egg_thrower_id IS NOT NULL AND egg_thrower_id = owner_id AND enquirer_id = ?
		ORDER BY egg_thrown_at DESC
	`
	var conversationIDs []int
	err := db.Select(&conversationIDs, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation IDs for user eggs: %w", err)
	}
	return conversationIDs, nil
}

// GetEggCountForUser returns the count of eggs bound to a user
// (eggs where owner threw at this user)
func GetEggCountForUser(userID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM conversations
		WHERE egg_thrower_id IS NOT NULL AND egg_thrower_id = owner_id AND enquirer_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get egg count for user: %w", err)
	}
	return count, nil
}

// GetConversationIDForUserEggByOrdinal returns the conversation ID at the given ordinal position (0-based)
// for eggs bound to a user (where egg_thrower_id = owner_id AND enquirer_id = userID), ordered by egg_thrown_at DESC
func GetConversationIDForUserEggByOrdinal(userID int, ordinal int) (int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE egg_thrower_id IS NOT NULL AND egg_thrower_id = owner_id AND enquirer_id = ?
		ORDER BY egg_thrown_at DESC
		LIMIT 1 OFFSET ?
	`
	var conversationID int
	err := db.QueryRow(query, userID, ordinal).Scan(&conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("conversation not found at ordinal %d for user %d", ordinal, userID)
		}
		return 0, fmt.Errorf("failed to get conversation ID for user egg by ordinal: %w", err)
	}
	return conversationID, nil
}

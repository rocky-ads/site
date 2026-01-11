package rock

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rocky-ads/site/db"
)

type Rock struct {
	ID             int       `db:"id" json:"id"`
	UserID         int       `db:"user_id" json:"user_id"`
	ConversationID int       `db:"conversation_id" json:"conversation_id"`
	ThrownAt       time.Time `db:"thrown_at" json:"thrown_at"`
}

var ErrRockNotFound = errors.New("rock not found")
var ErrMaxRocksReached = errors.New("user has reached maximum outstanding rocks")
var ErrRockAlreadyThrown = errors.New("a rock has already been thrown at this conversation")

// ThrowRock throws a rock at a conversation, making it public
func ThrowRock(userID, conversationID int) error {
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

	// Insert rock
	_, err = db.Exec(`
		INSERT INTO rocks (user_id, conversation_id, thrown_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`, userID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to insert rock: %w", err)
	}

	// Make conversation public
	_, err = db.Exec(`
		UPDATE conversations
		SET is_public = 1
		WHERE id = ?
	`, conversationID)
	if err != nil {
		return fmt.Errorf("failed to make conversation public: %w", err)
	}

	return nil
}

// UnthrowRock removes a rock from a conversation
func UnthrowRock(userID, conversationID int) error {
	// Delete rock
	result, err := db.Exec(`
		DELETE FROM rocks
		WHERE user_id = ? AND conversation_id = ?
	`, userID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to delete rock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrRockNotFound
	}

	// Check if conversation has any remaining rocks
	rockCount, err := GetRockCountForConversation(conversationID)
	if err != nil {
		return fmt.Errorf("failed to check remaining rocks: %w", err)
	}

	// If no rocks remain, make conversation private
	if rockCount == 0 {
		_, err = db.Exec(`
			UPDATE conversations
			SET is_public = 0
			WHERE id = ?
		`, conversationID)
		if err != nil {
			return fmt.Errorf("failed to make conversation private: %w", err)
		}
	}

	return nil
}

// GetRocksForAd returns all rocks for an ad (joins through conversations)
func GetRocksForAd(adID int) ([]Rock, error) {
	query := `
		SELECT r.id, r.user_id, r.conversation_id, r.thrown_at
		FROM rocks r
		JOIN conversations c ON r.conversation_id = c.id
		WHERE c.ad_id = ?
		ORDER BY r.thrown_at ASC
	`
	var rocks []Rock
	err := db.Select(&rocks, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rocks for ad: %w", err)
	}
	return rocks, nil
}

// GetPublicConversationsForAd returns public conversation IDs for an ad
func GetPublicConversationsForAd(adID int) ([]int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE ad_id = ? AND is_public = 1
		ORDER BY updated_at DESC
	`
	var conversationIDs []int
	err := db.Select(&conversationIDs, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public conversations for ad: %w", err)
	}
	return conversationIDs, nil
}

// GetPublicConversationIDByOrdinal returns the conversation ID at the given ordinal position (0-based)
// for public conversations for an ad, ordered by updated_at DESC
func GetPublicConversationIDByOrdinal(adID int, ordinal int) (int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE ad_id = ? AND is_public = 1
		ORDER BY updated_at DESC
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

// GetUserRockCount returns the count of outstanding rocks for a user
func GetUserRockCount(userID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM rocks
		WHERE user_id = ?
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user rock count: %w", err)
	}
	return count, nil
}

// GetRockCountForConversation returns the count of rocks for a conversation
func GetRockCountForConversation(conversationID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM rocks
		WHERE conversation_id = ?
	`, conversationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get rock count for conversation: %w", err)
	}
	return count, nil
}

// HasUserThrownRock checks if a user has thrown a rock at a conversation
func HasUserThrownRock(userID, conversationID int) (bool, error) {
	var exists int
	err := db.QueryRow(`
		SELECT 1
		FROM rocks
		WHERE user_id = ? AND conversation_id = ?
		LIMIT 1
	`, userID, conversationID).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, fmt.Errorf("failed to check if user threw rock: %w", err)
}

// GetRockForUserAndConversation gets a rock if it exists
func GetRockForUserAndConversation(userID, conversationID int) (Rock, error) {
	var rock Rock
	query := `
		SELECT id, user_id, conversation_id, thrown_at
		FROM rocks
		WHERE user_id = ? AND conversation_id = ?
	`
	err := db.QueryRow(query, userID, conversationID).Scan(
		&rock.ID,
		&rock.UserID,
		&rock.ConversationID,
		&rock.ThrownAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Rock{}, ErrRockNotFound
		}
		return Rock{}, fmt.Errorf("failed to get rock: %w", err)
	}
	return rock, nil
}

// GetRockForConversation gets the rock for a conversation (there should only be one)
func GetRockForConversation(conversationID int) (Rock, error) {
	var rock Rock
	query := `
		SELECT id, user_id, conversation_id, thrown_at
		FROM rocks
		WHERE conversation_id = ?
		LIMIT 1
	`
	err := db.QueryRow(query, conversationID).Scan(
		&rock.ID,
		&rock.UserID,
		&rock.ConversationID,
		&rock.ThrownAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Rock{}, ErrRockNotFound
		}
		return Rock{}, fmt.Errorf("failed to get rock: %w", err)
	}
	return rock, nil
}

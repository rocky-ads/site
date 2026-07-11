package rock

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/journal"
)

var ErrRockNotFound = errors.New("rock not found")
var ErrMaxRocksReached = errors.New("user has reached maximum outstanding rocks")
var ErrRockAlreadyThrown = errors.New("a rock has already been thrown at this conversation")

// ThrowRock throws a rock at a conversation, making it public
// If inquirer throws: rock_thrower_id = inquirer_id (bound to ad)
// If owner throws: rock_thrower_id = owner_id (bound to inquirer)
func ThrowRock(userID, conversationID int) error {
	conv, err := getConversationForRock(conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	if conv.OwnerID != userID && conv.InquirerID != userID {
		return fmt.Errorf("only conversation participants can throw rocks")
	}

	count, err := GetUserRockCount(userID)
	if err != nil {
		return fmt.Errorf("failed to check rock count: %w", err)
	}
	if count >= config.MaxOutstandingRocks {
		return ErrMaxRocksReached
	}

	rockCount, err := GetRockCountForConversation(conversationID)
	if err != nil {
		return fmt.Errorf("failed to check existing rocks: %w", err)
	}
	if rockCount > 0 {
		return ErrRockAlreadyThrown
	}

	now := time.Now().UTC()
	newJournal := journal.AppendRock(conv.Journal, journal.RockThrown, userID,
		now, time.UTC)

	_, err = db.Exec(`
		UPDATE conversations
		SET rock_thrower_id = $1, rock_thrown_at = $2,
			journal = $3, updated_at = $2
		WHERE id = $4
	`, userID, now, newJournal, conversationID)
	if err != nil {
		return fmt.Errorf("failed to throw rock: %w", err)
	}

	return nil
}

func getConversationForRock(conversationID int) (struct {
	OwnerID    int
	InquirerID int
	Journal    string
}, error) {
	var conv struct {
		OwnerID    int
		InquirerID int
		Journal    string
	}
	err := db.QueryRow(`
		SELECT owner_id, inquirer_id, journal
		FROM conversations
		WHERE id = $1
	`, conversationID).Scan(&conv.OwnerID, &conv.InquirerID, &conv.Journal)
	if err != nil {
		return conv, fmt.Errorf("failed to get conversation: %w", err)
	}
	return conv, nil
}

// UnthrowRock removes a rock from a conversation
func UnthrowRock(userID, conversationID int) error {
	var rockThrowerID sql.NullInt64
	var j string
	err := db.QueryRow(`
		SELECT rock_thrower_id, journal
		FROM conversations
		WHERE id = $1
	`, conversationID).Scan(&rockThrowerID, &j)
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

	now := time.Now().UTC()
	newJournal := journal.AppendRock(j, journal.RockUnthrown, userID, now,
		time.UTC)

	_, err = db.Exec(`
		UPDATE conversations
		SET rock_thrower_id = NULL, rock_thrown_at = NULL,
			journal = $1, updated_at = $2
		WHERE id = $3
	`, newJournal, now, conversationID)
	if err != nil {
		return fmt.Errorf("failed to remove rock: %w", err)
	}

	return nil
}

// GetPublicConversationsForAd returns public conversation IDs for an ad
// Only returns conversations with rocks bound to the ad (rock_thrower_id = inquirer_id)
func GetPublicConversationsForAd(adID int) ([]int, error) {
	query := `
		SELECT c.id
		FROM conversations c
		WHERE c.ad_id = $1 AND c.rock_thrower_id IS NOT NULL AND c.rock_thrower_id = c.inquirer_id
		ORDER BY c.updated_at DESC
	`
	var conversationIDs []int
	err := db.Select(&conversationIDs, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public conversations for ad: %w", err)
	}
	return conversationIDs, nil
}

// GetPublicConversationIDByOrdinal returns the conversation ID at the given ordinal position (0-based)
// for public conversations with rocks bound to the ad (rock_thrower_id = inquirer_id), ordered by latest activity DESC
func GetPublicConversationIDByOrdinal(adID int, ordinal int) (int, error) {
	query := `
		SELECT c.id
		FROM conversations c
		WHERE c.ad_id = $1 AND c.rock_thrower_id IS NOT NULL AND c.rock_thrower_id = c.inquirer_id
		ORDER BY c.updated_at DESC
		LIMIT 1 OFFSET $2
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
		WHERE rock_thrower_id = $1
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
		WHERE id = $1
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
		WHERE id = $1
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
// (where rock_thrower_id = owner_id, meaning owner threw rock at inquirer)
func GetConversationIDsForUserRocks(userID int) ([]int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE rock_thrower_id IS NOT NULL AND rock_thrower_id = owner_id AND inquirer_id = $1
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
		WHERE rock_thrower_id IS NOT NULL AND rock_thrower_id = owner_id AND inquirer_id = $1
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get rock count for user: %w", err)
	}
	return count, nil
}

// GetConversationIDForUserRockByOrdinal returns the conversation ID at the given ordinal position (0-based)
// for rocks bound to a user (where rock_thrower_id = owner_id AND inquirer_id = userID), ordered by rock_thrown_at DESC
func GetConversationIDForUserRockByOrdinal(userID int, ordinal int) (int, error) {
	query := `
		SELECT id
		FROM conversations
		WHERE rock_thrower_id IS NOT NULL AND rock_thrower_id = owner_id AND inquirer_id = $1
		ORDER BY rock_thrown_at DESC
		LIMIT 1 OFFSET $2
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

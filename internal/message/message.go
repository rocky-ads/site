package message

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/egg"
	"github.com/rocky-ads/site/internal/user"
)

type Conversation struct {
	ID                int        `db:"id"`
	AdID              int        `db:"ad_id"`
	OwnerID           int        `db:"owner_id"`
	EnquirerID        int        `db:"enquirer_id"`
	CreatedAt         time.Time  `db:"created_at"` // Computed: MIN(messages.created_at) or egg_thrown_at
	UpdatedAt         time.Time  `db:"updated_at"` // Computed: MAX(MAX(messages.created_at), egg_thrown_at)
	OwnerHasUnread    bool       `db:"owner_has_unread"`
	EnquirerHasUnread bool       `db:"enquirer_has_unread"`
	EggThrowerID      *int       `db:"egg_thrower_id"` // nil = no egg (private), NOT NULL = public, owner_id = bound to enquirer, enquirer_id = bound to ad
	EggThrownAt       *time.Time `db:"egg_thrown_at"`  // Only valid if egg_thrower_id IS NOT NULL
}

type Message struct {
	ID             int       `db:"id"`
	ConversationID int       `db:"conversation_id"`
	SenderID       int       `db:"sender_id"`
	Content        string    `db:"content"`
	CreatedAt      time.Time `db:"created_at"`
}

type ConversationWithAd struct {
	Conversation
	AdTitle string `db:"ad_title"`
}

type ConversationWithLastMessage struct {
	ConversationWithAd
	LastMessageContent string     `db:"last_message_content"`
	LastMessageAt      *time.Time `db:"last_message_at"`
	OtherUserID        int        `db:"other_user_id"`
	HasUnread          bool       `db:"has_unread"`
	RockCount          int        `db:"rock_count"`
	OtherUserName      string
	OtherUserEggCount  int
}

var ErrConversationNotFound = errors.New("conversation not found")
var ErrNotParticipant = errors.New("user is not a participant in this conversation")

// GetConversationByAdAndEnquirer gets an existing conversation by ad ID and enquirer ID
// Returns ErrConversationNotFound if the conversation does not exist
func GetConversationByAdAndEnquirer(adID, ownerID, enquirerID int) (Conversation, error) {
	if ownerID == enquirerID {
		return Conversation{}, fmt.Errorf("owner and enquirer cannot be the same")
	}

	var conv Conversation
	var eggThrownAt sql.NullTime
	query := `
		SELECT
			c.id, c.ad_id, c.owner_id, c.enquirer_id,
			c.owner_has_unread, c.enquirer_has_unread, c.egg_thrower_id, c.egg_thrown_at,
			COALESCE(MIN(m.created_at), c.egg_thrown_at, datetime('now')) AS created_at,
			COALESCE((SELECT MAX(ts) FROM (
				SELECT MAX(m2.created_at) AS ts FROM messages m2 WHERE m2.conversation_id = c.id
				UNION ALL
				SELECT c.egg_thrown_at AS ts WHERE c.egg_thrown_at IS NOT NULL
			)), datetime('now')) AS updated_at
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.ad_id = ? AND c.enquirer_id = ?
		GROUP BY c.id
	`
	var eggThrowerID sql.NullInt64
	var createdAtStr string
	var updatedAtStr string
	err := db.QueryRow(query, adID, enquirerID).Scan(
		&conv.ID,
		&conv.AdID,
		&conv.OwnerID,
		&conv.EnquirerID,
		&conv.OwnerHasUnread,
		&conv.EnquirerHasUnread,
		&eggThrowerID,
		&eggThrownAt,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, fmt.Errorf("failed to query conversation: %w", err)
	}

	// Parse timestamps from strings
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		// Try with microseconds if present
		createdAt, err = time.Parse("2006-01-02 15:04:05.999999999", createdAtStr)
		if err != nil {
			createdAt = time.Now()
		}
	}
	conv.CreatedAt = createdAt

	updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr)
	if err != nil {
		// Try with microseconds if present
		updatedAt, err = time.Parse("2006-01-02 15:04:05.999999999", updatedAtStr)
		if err != nil {
			updatedAt = time.Now()
		}
	}
	conv.UpdatedAt = updatedAt
	if eggThrowerID.Valid {
		id := int(eggThrowerID.Int64)
		conv.EggThrowerID = &id
	} else {
		conv.EggThrowerID = nil
	}
	if eggThrownAt.Valid {
		conv.EggThrownAt = &eggThrownAt.Time
	} else {
		conv.EggThrownAt = nil
	}
	// Verify owner_id matches (in case ad ownership changed)
	if conv.OwnerID != ownerID {
		// Update owner_id if it changed
		_, updateErr := db.Exec(`
			UPDATE conversations
			SET owner_id = ?
			WHERE id = ?
		`, ownerID, conv.ID)
		if updateErr != nil {
			return Conversation{}, fmt.Errorf("failed to update conversation owner: %w", updateErr)
		}
		conv.OwnerID = ownerID
	}
	return conv, nil
}

// CreateConversation creates a new conversation
func CreateConversation(adID, ownerID, enquirerID int) (Conversation, error) {
	if ownerID == enquirerID {
		return Conversation{}, fmt.Errorf("owner and enquirer cannot be the same")
	}

	result, err := db.Exec(`
		INSERT INTO conversations (ad_id, owner_id, enquirer_id, owner_has_unread, enquirer_has_unread, egg_thrower_id)
		VALUES (?, ?, ?, 0, 0, NULL)
	`, adID, ownerID, enquirerID)
	if err != nil {
		// If insert failed due to unique constraint (race condition), try to get the existing conversation
		errStr := err.Error()
		if errStr == "UNIQUE constraint failed: conversations.ad_id, conversations.enquirer_id" ||
			errStr == "constraint failed: conversations.ad_id, conversations.enquirer_id" ||
			errStr == "UNIQUE constraint failed" {
			// Retry SELECT to get the conversation that was just created by another request
			return GetConversationByAdAndEnquirer(adID, ownerID, enquirerID)
		}
		return Conversation{}, fmt.Errorf("failed to create conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to get conversation ID: %w", err)
	}

	var conv Conversation
	conv.ID = int(id)
	conv.AdID = adID
	conv.OwnerID = ownerID
	conv.EnquirerID = enquirerID
	conv.EggThrowerID = nil
	conv.EggThrownAt = nil
	// CreatedAt and UpdatedAt will be computed from messages/egg_thrown_at on next query

	return conv, nil
}

func GetConversation(conversationID, userID int) (Conversation, error) {
	var conv Conversation
	var eggThrownAt sql.NullTime
	query := `
		SELECT 
			c.id, c.ad_id, c.owner_id, c.enquirer_id,
			c.owner_has_unread, c.enquirer_has_unread, c.egg_thrower_id, c.egg_thrown_at,
			COALESCE(MIN(m.created_at), c.egg_thrown_at, datetime('now')) AS created_at,
			COALESCE((SELECT MAX(ts) FROM (
				SELECT MAX(m2.created_at) AS ts FROM messages m2 WHERE m2.conversation_id = c.id
				UNION ALL
				SELECT c.egg_thrown_at AS ts WHERE c.egg_thrown_at IS NOT NULL
			)), datetime('now')) AS updated_at
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.id = ? AND (c.owner_id = ? OR c.enquirer_id = ? OR c.egg_thrower_id IS NOT NULL)
		GROUP BY c.id
	`
	var eggThrowerID sql.NullInt64
	var createdAtStr string
	var updatedAtStr string
	err := db.QueryRow(query, conversationID, userID, userID).Scan(
		&conv.ID,
		&conv.AdID,
		&conv.OwnerID,
		&conv.EnquirerID,
		&conv.OwnerHasUnread,
		&conv.EnquirerHasUnread,
		&eggThrowerID,
		&eggThrownAt,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Parse timestamps from strings
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		// Try with microseconds if present
		createdAt, err = time.Parse("2006-01-02 15:04:05.999999999", createdAtStr)
		if err != nil {
			createdAt = time.Now()
		}
	}
	conv.CreatedAt = createdAt

	updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr)
	if err != nil {
		// Try with microseconds if present
		updatedAt, err = time.Parse("2006-01-02 15:04:05.999999999", updatedAtStr)
		if err != nil {
			updatedAt = time.Now()
		}
	}
	conv.UpdatedAt = updatedAt
	if eggThrowerID.Valid {
		id := int(eggThrowerID.Int64)
		conv.EggThrowerID = &id
	} else {
		conv.EggThrowerID = nil
	}
	if eggThrownAt.Valid {
		conv.EggThrownAt = &eggThrownAt.Time
	} else {
		conv.EggThrownAt = nil
	}
	return conv, nil
}

// GetConversationByID gets a conversation by ID without user check (for public access)
func GetConversationByID(conversationID int) (Conversation, error) {
	var conv Conversation
	var eggThrownAt sql.NullTime
	query := `
		SELECT
			c.id, c.ad_id, c.owner_id, c.enquirer_id,
			c.owner_has_unread, c.enquirer_has_unread, c.egg_thrower_id, c.egg_thrown_at,
			COALESCE(MIN(m.created_at), c.egg_thrown_at, datetime('now')) AS created_at,
			COALESCE((SELECT MAX(ts) FROM (
				SELECT MAX(m2.created_at) AS ts FROM messages m2 WHERE m2.conversation_id = c.id
				UNION ALL
				SELECT c.egg_thrown_at AS ts WHERE c.egg_thrown_at IS NOT NULL
			)), datetime('now')) AS updated_at
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.id = ?
		GROUP BY c.id
	`
	var eggThrowerID sql.NullInt64
	var createdAtStr string
	var updatedAtStr string
	err := db.QueryRow(query, conversationID).Scan(
		&conv.ID,
		&conv.AdID,
		&conv.OwnerID,
		&conv.EnquirerID,
		&conv.OwnerHasUnread,
		&conv.EnquirerHasUnread,
		&eggThrowerID,
		&eggThrownAt,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Parse timestamps from strings
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		// Try with microseconds if present
		createdAt, err = time.Parse("2006-01-02 15:04:05.999999999", createdAtStr)
		if err != nil {
			createdAt = time.Now()
		}
	}
	conv.CreatedAt = createdAt

	updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr)
	if err != nil {
		// Try with microseconds if present
		updatedAt, err = time.Parse("2006-01-02 15:04:05.999999999", updatedAtStr)
		if err != nil {
			updatedAt = time.Now()
		}
	}
	conv.UpdatedAt = updatedAt

	if eggThrowerID.Valid {
		id := int(eggThrowerID.Int64)
		conv.EggThrowerID = &id
	} else {
		conv.EggThrowerID = nil
	}
	if eggThrownAt.Valid {
		conv.EggThrownAt = &eggThrownAt.Time
	} else {
		conv.EggThrownAt = nil
	}
	return conv, nil
}

// IsPublic checks if a conversation is public (has an egg thrown)
func IsPublic(conversationID int) (bool, error) {
	var eggThrowerID sql.NullInt64
	err := db.QueryRow(`
		SELECT egg_thrower_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&eggThrowerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrConversationNotFound
		}
		return false, fmt.Errorf("failed to check if conversation is public: %w", err)
	}
	return eggThrowerID.Valid, nil
}

// CanUserPost checks if a user can post to a conversation (must be participant)
func CanUserPost(conversationID, userID int) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM conversations
		WHERE id = ? AND (owner_id = ? OR enquirer_id = ?)
	`, conversationID, userID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if user can post: %w", err)
	}
	return count > 0, nil
}

// GetPublicConversations returns public conversations for an ad
func GetPublicConversations(adID int) ([]Conversation, error) {
	query := `
		SELECT 
			c.id, c.ad_id, c.owner_id, c.enquirer_id,
			c.owner_has_unread, c.enquirer_has_unread, c.egg_thrower_id, c.egg_thrown_at,
			COALESCE(MIN(m.created_at), c.egg_thrown_at, datetime('now')) AS created_at,
			COALESCE(
				(SELECT MAX(ts) FROM (
					SELECT MAX(m2.created_at) AS ts FROM messages m2 WHERE m2.conversation_id = c.id
					UNION ALL
					SELECT c.egg_thrown_at AS ts WHERE c.egg_thrown_at IS NOT NULL
				)),
				datetime('now')
			) AS updated_at
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		WHERE c.ad_id = ? AND c.egg_thrower_id IS NOT NULL
		GROUP BY c.id
		ORDER BY updated_at DESC
	`
	// Use Query instead of Select to manually handle datetime string conversion
	rows, err := db.Query(query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public conversations: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		var eggThrowerID sql.NullInt64
		var eggThrownAt sql.NullTime
		var createdAtStr string
		var updatedAtStr string

		err := rows.Scan(
			&conv.ID,
			&conv.AdID,
			&conv.OwnerID,
			&conv.EnquirerID,
			&conv.OwnerHasUnread,
			&conv.EnquirerHasUnread,
			&eggThrowerID,
			&eggThrownAt,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		// Parse timestamps from strings
		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			// Try with microseconds if present
			createdAt, err = time.Parse("2006-01-02 15:04:05.999999999", createdAtStr)
			if err != nil {
				createdAt = time.Now()
			}
		}
		conv.CreatedAt = createdAt

		updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if err != nil {
			// Try with microseconds if present
			updatedAt, err = time.Parse("2006-01-02 15:04:05.999999999", updatedAtStr)
			if err != nil {
				updatedAt = time.Now()
			}
		}
		conv.UpdatedAt = updatedAt

		if eggThrowerID.Valid {
			id := int(eggThrowerID.Int64)
			conv.EggThrowerID = &id
		}
		if eggThrownAt.Valid {
			conv.EggThrownAt = &eggThrownAt.Time
		}

		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate conversations: %w", err)
	}

	return conversations, nil
}

func GetUserConversations(userID int, loc *time.Location) ([]ConversationWithLastMessage, error) {
	query := `
		SELECT
			c.id,
			c.ad_id,
			c.owner_id,
			c.enquirer_id,
			c.owner_has_unread,
			c.enquirer_has_unread,
			c.egg_thrower_id,
			c.egg_thrown_at,
			a.title AS ad_title,
			COALESCE(m.content, '') AS last_message_content,
			m.created_at AS last_message_at,
			COALESCE((
				SELECT COUNT(*)
				FROM conversations c2
				WHERE c2.ad_id = a.id
					AND c2.egg_thrower_id IS NOT NULL
					AND c2.egg_thrower_id = c2.enquirer_id
			), 0) AS rock_count,
			COALESCE(MIN(msg.created_at), c.egg_thrown_at, datetime('now')) AS created_at,
			COALESCE(
				(SELECT MAX(ts) FROM (
					SELECT MAX(msg2.created_at) AS ts FROM messages msg2 WHERE msg2.conversation_id = c.id
					UNION ALL
					SELECT c.egg_thrown_at AS ts WHERE c.egg_thrown_at IS NOT NULL
				)),
				datetime('now')
			) AS updated_at,
			CASE
				WHEN c.owner_id = ? THEN c.enquirer_id
				ELSE c.owner_id
			END AS other_user_id,
			CASE
				WHEN c.owner_id = ? THEN c.owner_has_unread
				ELSE c.enquirer_has_unread
			END AS has_unread
		FROM conversations c
		JOIN ads a ON c.ad_id = a.id
		LEFT JOIN (
			SELECT conversation_id, content, created_at
			FROM messages
			WHERE id IN (
				SELECT MAX(id) FROM messages GROUP BY conversation_id
			)
		) m ON c.id = m.conversation_id
		LEFT JOIN messages msg ON c.id = msg.conversation_id
		WHERE (c.owner_id = ? OR c.enquirer_id = ?)
		GROUP BY c.id, a.title, m.content, m.created_at
		ORDER BY updated_at DESC
	`
	// Use Query instead of Select to manually handle datetime string conversion
	rows, err := db.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}
	defer rows.Close()

	var conversations []ConversationWithLastMessage
	for rows.Next() {
		var conv ConversationWithLastMessage
		var eggThrowerID sql.NullInt64
		var eggThrownAt sql.NullTime
		var lastMessageAt sql.NullTime
		var createdAtStr string
		var updatedAtStr string

		err := rows.Scan(
			&conv.ID,
			&conv.AdID,
			&conv.OwnerID,
			&conv.EnquirerID,
			&conv.OwnerHasUnread,
			&conv.EnquirerHasUnread,
			&eggThrowerID,
			&eggThrownAt,
			&conv.AdTitle,
			&conv.LastMessageContent,
			&lastMessageAt,
			&conv.RockCount,
			&createdAtStr,
			&updatedAtStr,
			&conv.OtherUserID,
			&conv.HasUnread,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		// Parse timestamps from strings
		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			// Try with microseconds if present
			createdAt, err = time.Parse("2006-01-02 15:04:05.999999999", createdAtStr)
			if err != nil {
				createdAt = time.Now()
			}
		}
		conv.CreatedAt = createdAt

		updatedAt, err := time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if err != nil {
			// Try with microseconds if present
			updatedAt, err = time.Parse("2006-01-02 15:04:05.999999999", updatedAtStr)
			if err != nil {
				updatedAt = time.Now()
			}
		}
		conv.UpdatedAt = updatedAt

		if eggThrowerID.Valid {
			id := int(eggThrowerID.Int64)
			conv.EggThrowerID = &id
		}
		if eggThrownAt.Valid {
			conv.EggThrownAt = &eggThrownAt.Time
		}
		if lastMessageAt.Valid {
			conv.LastMessageAt = &lastMessageAt.Time
		}

		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate conversations: %w", err)
	}

	for i := range conversations {
		conversations[i].CreatedAt = conversations[i].CreatedAt.In(loc)
		conversations[i].UpdatedAt = conversations[i].UpdatedAt.In(loc)
		if conversations[i].LastMessageAt != nil {
			converted := (*conversations[i].LastMessageAt).In(loc)
			conversations[i].LastMessageAt = &converted
		}
	}

	enrichConversationListItems(conversations)

	return conversations, nil
}

func enrichConversationListItems(conversations []ConversationWithLastMessage) {
	if len(conversations) == 0 {
		return
	}

	otherIDs := make(map[int]struct{})
	for _, conv := range conversations {
		otherIDs[conv.OtherUserID] = struct{}{}
	}

	names := make(map[int]string, len(otherIDs))
	eggCounts := make(map[int]int, len(otherIDs))
	for id := range otherIDs {
		if u, err := user.GetByID(id); err == nil {
			names[id] = u.Name
		}
		count, err := egg.GetEggCountForUser(id)
		if err == nil {
			eggCounts[id] = count
		}
	}

	for i := range conversations {
		id := conversations[i].OtherUserID
		if name, ok := names[id]; ok {
			conversations[i].OtherUserName = name
		} else {
			conversations[i].OtherUserName = "Unknown User"
		}
		conversations[i].OtherUserEggCount = eggCounts[id]
	}
}

func GetConversationMessages(conversationID, userID int, loc *time.Location) ([]Message, error) {
	// Allow access if user is participant OR conversation is public
	conv, err := GetConversation(conversationID, userID)
	if err != nil {
		return nil, err
	}
	_ = conv

	query := `
		SELECT id, conversation_id, sender_id, content, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`
	var messages []Message
	err = db.Select(&messages, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	for i := range messages {
		messages[i].CreatedAt = messages[i].CreatedAt.In(loc)
	}

	return messages, nil
}

func CreateMessage(conversationID, senderID int, content string) (Message, error) {
	if content == "" {
		return Message{}, fmt.Errorf("message content cannot be empty")
	}

	// Get conversation (or create if it doesn't exist - this handles the case where conversationID is 0)
	var conv Conversation
	var err error
	if conversationID == 0 {
		// This shouldn't happen in normal flow, but handle it gracefully
		return Message{}, fmt.Errorf("conversation ID is required")
	}

	conv, err = GetConversationByID(conversationID)
	if err != nil {
		return Message{}, err
	}

	// Check if user can post (must be participant)
	canPost, err := CanUserPost(conversationID, senderID)
	if err != nil {
		return Message{}, err
	}
	if !canPost {
		return Message{}, ErrNotParticipant
	}

	result, err := db.Exec(`
		INSERT INTO messages (conversation_id, sender_id, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, conversationID, senderID, content)
	if err != nil {
		return Message{}, fmt.Errorf("failed to create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("failed to get message ID: %w", err)
	}

	// Set has_unread=true for the recipient (the one who didn't send the message)
	var recipientField string
	if conv.OwnerID == senderID {
		recipientField = "enquirer_has_unread"
	} else {
		recipientField = "owner_has_unread"
	}

	_, err = db.Exec(fmt.Sprintf(`
		UPDATE conversations
		SET %s = 1
		WHERE id = ?
	`, recipientField), conversationID)
	if err != nil {
		return Message{}, fmt.Errorf("failed to update conversation: %w", err)
	}

	var msg Message
	query := `
		SELECT id, conversation_id, sender_id, content, created_at
		FROM messages
		WHERE id = ?
	`
	err = db.QueryRow(query, id).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.Content,
		&msg.CreatedAt,
	)
	if err != nil {
		return Message{}, fmt.Errorf("failed to get created message: %w", err)
	}

	return msg, nil
}

// GetHasUnread checks if the user has any unread conversations
func GetHasUnread(userID int) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM conversations
		WHERE (owner_id = ? AND owner_has_unread = 1)
		   OR (enquirer_id = ? AND enquirer_has_unread = 1)
	`
	var hasUnread bool
	err := db.QueryRow(query, userID, userID).Scan(&hasUnread)
	if err != nil {
		return false, fmt.Errorf("failed to get unread status: %w", err)
	}
	return hasUnread, nil
}

// MarkConversationAsRead marks a conversation as read for the given user
func MarkConversationAsRead(conversationID, userID int) error {
	conv, err := GetConversation(conversationID, userID)
	if err != nil {
		return err
	}

	var field string
	if conv.OwnerID == userID {
		field = "owner_has_unread"
	} else {
		field = "enquirer_has_unread"
	}

	_, err = db.Exec(fmt.Sprintf(`
		UPDATE conversations
		SET %s = 0
		WHERE id = ?
	`, field), conversationID)
	if err != nil {
		return fmt.Errorf("failed to mark conversation as read: %w", err)
	}

	return nil
}

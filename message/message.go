package message

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rocky-ads/site/db"
)

type Conversation struct {
	ID                int       `db:"id" json:"id"`
	AdID              int       `db:"ad_id" json:"ad_id"`
	OwnerID           int       `db:"owner_id" json:"owner_id"`
	EnquirerID        int       `db:"enquirer_id" json:"enquirer_id"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
	OwnerHasUnread    bool      `db:"owner_has_unread" json:"owner_has_unread"`
	EnquirerHasUnread bool      `db:"enquirer_has_unread" json:"enquirer_has_unread"`
	IsPublic          bool      `db:"is_public" json:"is_public"`
}

type Message struct {
	ID             int       `db:"id" json:"id"`
	ConversationID int       `db:"conversation_id" json:"conversation_id"`
	SenderID       int       `db:"sender_id" json:"sender_id"`
	Content        string    `db:"content" json:"content"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type ConversationWithAd struct {
	Conversation
	AdTitle string `db:"ad_title" json:"ad_title"`
}

type ConversationWithLastMessage struct {
	ConversationWithAd
	LastMessageContent string     `db:"last_message_content" json:"last_message_content"`
	LastMessageAt      *time.Time `db:"last_message_at" json:"last_message_at"`
	OtherUserID        int        `db:"other_user_id" json:"other_user_id"`
	HasUnread          bool       `db:"has_unread" json:"has_unread"`
}

var ErrConversationNotFound = errors.New("conversation not found")
var ErrNotParticipant = errors.New("user is not a participant in this conversation")

func GetOrCreateConversation(adID, ownerID, enquirerID int) (Conversation, error) {
	if ownerID == enquirerID {
		return Conversation{}, fmt.Errorf("owner and enquirer cannot be the same")
	}

	var conv Conversation
	query := `
		SELECT id, ad_id, owner_id, enquirer_id, created_at, updated_at, owner_has_unread, enquirer_has_unread, is_public
		FROM conversations
		WHERE ad_id = ? AND enquirer_id = ?
	`
	err := db.QueryRow(query, adID, enquirerID).Scan(
		&conv.ID,
		&conv.AdID,
		&conv.OwnerID,
		&conv.EnquirerID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
		&conv.OwnerHasUnread,
		&conv.EnquirerHasUnread,
		&conv.IsPublic,
	)
	if err == nil {
		return conv, nil
	}
	if err != sql.ErrNoRows {
		return Conversation{}, fmt.Errorf("failed to query conversation: %w", err)
	}

	result, err := db.Exec(`
		INSERT INTO conversations (ad_id, owner_id, enquirer_id, created_at, updated_at, owner_has_unread, enquirer_has_unread, is_public)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0, 0, 0)
	`, adID, ownerID, enquirerID)
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to create conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to get conversation ID: %w", err)
	}

	conv.ID = int(id)
	conv.AdID = adID
	conv.OwnerID = ownerID
	conv.EnquirerID = enquirerID
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = time.Now()

	return conv, nil
}

func GetConversation(conversationID, userID int) (Conversation, error) {
	var conv Conversation
	query := `
		SELECT id, ad_id, owner_id, enquirer_id, created_at, updated_at, owner_has_unread, enquirer_has_unread, is_public
		FROM conversations
		WHERE id = ? AND (owner_id = ? OR enquirer_id = ? OR is_public = 1)
	`
	err := db.QueryRow(query, conversationID, userID, userID).Scan(
		&conv.ID,
		&conv.AdID,
		&conv.OwnerID,
		&conv.EnquirerID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
		&conv.OwnerHasUnread,
		&conv.EnquirerHasUnread,
		&conv.IsPublic,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, fmt.Errorf("failed to get conversation: %w", err)
	}
	return conv, nil
}

// GetConversationByID gets a conversation by ID without user check (for public access)
func GetConversationByID(conversationID int) (Conversation, error) {
	var conv Conversation
	query := `
		SELECT id, ad_id, owner_id, enquirer_id, created_at, updated_at, owner_has_unread, enquirer_has_unread, is_public
		FROM conversations
		WHERE id = ?
	`
	err := db.QueryRow(query, conversationID).Scan(
		&conv.ID,
		&conv.AdID,
		&conv.OwnerID,
		&conv.EnquirerID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
		&conv.OwnerHasUnread,
		&conv.EnquirerHasUnread,
		&conv.IsPublic,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, fmt.Errorf("failed to get conversation: %w", err)
	}
	return conv, nil
}

// IsPublic checks if a conversation is public
func IsPublic(conversationID int) (bool, error) {
	var isPublic int
	err := db.QueryRow(`
		SELECT is_public
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&isPublic)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrConversationNotFound
		}
		return false, fmt.Errorf("failed to check if conversation is public: %w", err)
	}
	return isPublic == 1, nil
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
		SELECT id, ad_id, owner_id, enquirer_id, created_at, updated_at, owner_has_unread, enquirer_has_unread, is_public
		FROM conversations
		WHERE ad_id = ? AND is_public = 1
		ORDER BY updated_at DESC
	`
	var conversations []Conversation
	err := db.Select(&conversations, query, adID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public conversations: %w", err)
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
			c.created_at,
			c.updated_at,
			c.owner_has_unread,
			c.enquirer_has_unread,
			c.is_public,
			a.title AS ad_title,
			COALESCE(m.content, '') AS last_message_content,
			m.created_at AS last_message_at,
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
		WHERE (c.owner_id = ? OR c.enquirer_id = ?)
		ORDER BY COALESCE(m.created_at, c.updated_at) DESC
	`
	var conversations []ConversationWithLastMessage
	err := db.Select(&conversations, query, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}

	for i := range conversations {
		conversations[i].CreatedAt = conversations[i].CreatedAt.In(loc)
		conversations[i].UpdatedAt = conversations[i].UpdatedAt.In(loc)
		if conversations[i].LastMessageAt != nil {
			converted := (*conversations[i].LastMessageAt).In(loc)
			conversations[i].LastMessageAt = &converted
		}
	}

	return conversations, nil
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

	// Check if user can post (must be participant)
	canPost, err := CanUserPost(conversationID, senderID)
	if err != nil {
		return Message{}, err
	}
	if !canPost {
		return Message{}, ErrNotParticipant
	}

	conv, err := GetConversationByID(conversationID)
	if err != nil {
		return Message{}, err
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
		SET updated_at = CURRENT_TIMESTAMP, %s = 1
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

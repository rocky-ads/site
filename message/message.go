package message

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rocky-ads/site/db"
)

type Conversation struct {
	ID         int       `db:"id" json:"id"`
	AdID       int       `db:"ad_id" json:"ad_id"`
	OwnerID    int       `db:"owner_id" json:"owner_id"`
	EnquirerID int       `db:"enquirer_id" json:"enquirer_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
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
}

var ErrConversationNotFound = errors.New("conversation not found")
var ErrNotParticipant = errors.New("user is not a participant in this conversation")

func GetOrCreateConversation(adID, ownerID, enquirerID int) (Conversation, error) {
	if ownerID == enquirerID {
		return Conversation{}, fmt.Errorf("owner and enquirer cannot be the same")
	}

	var conv Conversation
	query := `
		SELECT id, ad_id, owner_id, enquirer_id, created_at, updated_at
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
	)
	if err == nil {
		return conv, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return Conversation{}, fmt.Errorf("failed to query conversation: %w", err)
	}

	result, err := db.Exec(`
		INSERT INTO conversations (ad_id, owner_id, enquirer_id, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
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
		SELECT id, ad_id, owner_id, enquirer_id, created_at, updated_at
		FROM conversations
		WHERE id = ? AND (owner_id = ? OR enquirer_id = ?)
	`
	err := db.QueryRow(query, conversationID, userID, userID).Scan(
		&conv.ID,
		&conv.AdID,
		&conv.OwnerID,
		&conv.EnquirerID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Conversation{}, ErrConversationNotFound
		}
		return Conversation{}, fmt.Errorf("failed to get conversation: %w", err)
	}
	return conv, nil
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
			a.title AS ad_title,
			COALESCE(m.content, '') AS last_message_content,
			m.created_at AS last_message_at,
			CASE
				WHEN c.owner_id = ? THEN c.enquirer_id
				ELSE c.owner_id
			END AS other_user_id
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
	err := db.Select(&conversations, query, userID, userID, userID)
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

	conv, err := GetConversation(conversationID, senderID)
	if err != nil {
		return Message{}, err
	}

	if conv.OwnerID != senderID && conv.EnquirerID != senderID {
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

	_, err = db.Exec(`
		UPDATE conversations
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, conversationID)
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

func GetUnreadMessageCount(userID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		JOIN conversations c ON m.conversation_id = c.id
		WHERE (c.owner_id = ? OR c.enquirer_id = ?)
		AND m.sender_id != ?
	`
	var count int
	err := db.QueryRow(query, userID, userID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

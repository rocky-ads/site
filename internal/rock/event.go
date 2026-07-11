package rock

import (
	"fmt"
	"time"

	"github.com/rocky-ads/site/internal/db"
)

const (
	EventThrown   = "thrown"
	EventUnthrown = "unthrown"
)

type Event struct {
	ID             int       `db:"id"`
	ConversationID int       `db:"conversation_id"`
	UserID         int       `db:"user_id"`
	Kind           string    `db:"kind"`
	CreatedAt      time.Time `db:"created_at"`
}

func RecordEvent(conversationID, userID int, kind string) error {
	_, err := db.Exec(`
		INSERT INTO rock_events (conversation_id, user_id, kind)
		VALUES ($1, $2, $3)
	`, conversationID, userID, kind)
	if err != nil {
		return fmt.Errorf("failed to record rock event: %w", err)
	}
	return nil
}

func recordThrownEventAt(conversationID, userID int, thrownAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO rock_events (conversation_id, user_id, kind, created_at)
		VALUES ($1, $2, $3, $4)
	`, conversationID, userID, EventThrown, thrownAt)
	if err != nil {
		return fmt.Errorf("failed to record rock thrown event: %w", err)
	}
	return nil
}

func EnsureThrownEventRecorded(conversationID, userID int,
	thrownAt time.Time) error {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM rock_events
		WHERE conversation_id = $1 AND kind = $2
	`, conversationID, EventThrown).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check rock thrown events: %w", err)
	}
	if count > 0 {
		return nil
	}
	return recordThrownEventAt(conversationID, userID, thrownAt)
}

func GetEvents(conversationID int, rockThrowerID *int,
	rockThrownAt *time.Time) ([]Event, error) {
	var events []Event
	err := db.Select(&events, `
		SELECT id, conversation_id, user_id, kind, created_at
		FROM rock_events
		WHERE conversation_id = $1
		ORDER BY created_at ASC, id ASC
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rock events: %w", err)
	}
	if len(events) == 0 && rockThrowerID != nil && rockThrownAt != nil {
		events = append(events, Event{
			ConversationID: conversationID,
			UserID:         *rockThrowerID,
			Kind:           EventThrown,
			CreatedAt:      *rockThrownAt,
		})
	}
	return events, nil
}

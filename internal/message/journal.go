package message

import (
	"time"

	"github.com/rocky-ads/site/internal/journal"
)

const (
	JournalMessage      = journal.Message
	JournalRockThrown   = journal.RockThrown
	JournalRockUnthrown = journal.RockUnthrown
)

type JournalEntry = journal.Entry

func AppendJournalEntry(j, label, meta, body string, at time.Time,
	tz *time.Location) string {
	return journal.AppendEntry(j, label, meta, body, at, tz)
}

func AppendMessageEntry(j string, senderID int, body string, at time.Time,
	tz *time.Location) string {
	return journal.AppendMessage(j, senderID, body, at, tz)
}

func AppendRockEntry(j, kind string, userID int, at time.Time, tz *time.Location) string {
	return journal.AppendRock(j, kind, userID, at, tz)
}

func ParseJournal(j string) []JournalEntry {
	return journal.Parse(j)
}

func LastMessagePreview(j string) (content string, at time.Time, ok bool) {
	return journal.LastMessagePreview(j)
}

func FirstEntryAt(j string) (time.Time, bool) {
	return journal.FirstEntryAt(j)
}

func MessagesFromJournal(conversationID int, j string, tz *time.Location) []Message {
	entries := journal.Parse(j)
	var messages []Message
	id := 1
	for _, e := range entries {
		if e.Kind != journal.Message {
			continue
		}
		at := e.At
		if tz != nil {
			at = at.In(tz)
		}
		messages = append(messages, Message{
			ID:             id,
			ConversationID: conversationID,
			SenderID:       e.UserID,
			Content:        e.Body,
			CreatedAt:      at,
		})
		id++
	}
	return messages
}

// RockJournalEvent is a rock throw/unthrow timeline entry from the journal.
type RockJournalEvent struct {
	UserID    int
	Kind      string
	CreatedAt time.Time
}

func RockEventsFromJournal(j string, tz *time.Location) []RockJournalEvent {
	entries := journal.Parse(j)
	var events []RockJournalEvent
	for _, e := range entries {
		var kind string
		switch e.Kind {
		case journal.RockThrown:
			kind = "thrown"
		case journal.RockUnthrown:
			kind = "unthrown"
		default:
			continue
		}
		at := e.At
		if tz != nil {
			at = at.In(tz)
		}
		events = append(events, RockJournalEvent{
			UserID:    e.UserID,
			Kind:      kind,
			CreatedAt: at,
		})
	}
	return events
}

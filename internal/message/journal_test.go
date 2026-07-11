package message

import (
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/journal"
)

func TestMessagesAndRockEventsFromJournal(t *testing.T) {
	loc := time.UTC
	t1 := time.Date(2026, 7, 10, 22, 33, 0, 0, loc)
	t2 := time.Date(2026, 7, 10, 22, 35, 0, 0, loc)
	t3 := time.Date(2026, 7, 10, 22, 40, 0, 0, loc)

	j := journal.AppendMessage("", 12, "hello", t1, loc)
	j = journal.AppendRock(j, journal.RockThrown, 12, t2, loc)
	j = journal.AppendMessage(j, 34, "reply", t3, loc)

	msgs := MessagesFromJournal(7, j, loc)
	if len(msgs) != 2 {
		t.Fatalf("messages = %d, want 2", len(msgs))
	}
	if msgs[0].ConversationID != 7 || msgs[0].SenderID != 12 ||
		msgs[0].Content != "hello" {
		t.Errorf("first message: %+v", msgs[0])
	}
	if msgs[1].SenderID != 34 || msgs[1].Content != "reply" {
		t.Errorf("second message: %+v", msgs[1])
	}

	events := RockEventsFromJournal(j, loc)
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Kind != "thrown" || events[0].UserID != 12 {
		t.Errorf("event: %+v", events[0])
	}
}

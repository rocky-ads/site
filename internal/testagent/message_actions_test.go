package testagent

import (
	"testing"

	"github.com/rocky-ads/site/internal/browserclient"
)

func TestReplyReadyActionPostsWhenFormOpen(t *testing.T) {
	conv := browserclient.ConversationSnapshot{
		ID: 3,
		Messages: []browserclient.ConversationMessage{
			{Text: "hi", FromSelf: false},
		},
	}
	page := browserclient.PageAffordances{
		Conversation:          &conv,
		OpenConversationForms: []int{3},
	}
	act, ok := replyReadyAction(page, newInbox(), Persona{Name: "messenger"})
	if !ok || act.Action != "post_form" || act.Path != "/auth/conversation/3/send" {
		t.Fatalf("act %+v ok=%v", act, ok)
	}
	if act.Fields["content"] == "" {
		t.Fatal("expected generated content")
	}
}

func TestConversationClickLoopDetected(t *testing.T) {
	j := NewJournal()
	for i := 0; i < 3; i++ {
		j.Append(JournalEntry{Phase: PhaseAct, Action: "CLICK /auth/conversation/2"})
	}
	if !conversationClickLoopDetected(j.Entries(), "/auth/conversation/2") {
		t.Fatal("expected loop detected")
	}
}

func TestEnsureMessageFieldsFillsContent(t *testing.T) {
	act := ensureMessageFields(PlannedAction{
		Action: "post_form",
		Path:   "/auth/ad/942/send",
	}, Persona{Name: "messenger"})
	if act.Fields["content"] == "" {
		t.Fatal("expected content")
	}
}

package handler_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/ui"
)

func TestConversationMessagesAppendSwap(t *testing.T) {
	got := ui.ConversationMessagesAppendSwapForTest(5)
	want := "beforeend scroll:#conversation-5-messages:bottom"
	if got != want {
		t.Fatalf("swap = %q, want %q", got, want)
	}
}

func TestRenderMessageAppendOOB(t *testing.T) {
	msgData := ui.MessageItemData{
		SenderID: 2, CurrentUserID: 4,
		Content:   "Hi this is Bob",
		CreatedAt: time.Now().UTC(),
	}
	var buf bytes.Buffer
	node := ui.MessageItem(msgData, ui.ConversationMessageAppendOOB(5))
	if err := node.Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	t.Logf("HTML:\n%s", html)
	if !strings.Contains(html, `hx-swap-oob="beforeend:#conversation-5-messages"`) {
		t.Fatalf("missing oob attr: %s", html)
	}
	if strings.Contains(html, "scroll:bottom") {
		t.Fatalf("scroll:bottom must not appear in oob attr: %s", html)
	}
	if strings.Contains(html, "\n") {
		t.Log("WARNING: HTML contains newlines - may break SSE")
	}
}

func TestRenderRockEventAppendOOB(t *testing.T) {
	event := ui.RockEventData{
		ThrowerID: 4, CurrentUserID: 3,
		Kind:    ui.RockEventThrown,
		EventAt: time.Now().UTC(),
		OwnerID: 3, InquirerID: 4,
	}
	var buf bytes.Buffer
	node := ui.RockEventMessage(event, ui.ConversationMessageAppendOOB(5))
	if err := node.Render(&buf); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	if !strings.Contains(html, `hx-swap-oob="beforeend:#conversation-5-messages"`) {
		t.Fatalf("missing oob attr: %s", html)
	}
	if !strings.Contains(html, "Rock thrown") {
		t.Fatalf("missing rock event label: %s", html)
	}
}

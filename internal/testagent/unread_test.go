package testagent

import (
	"testing"

	"github.com/rocky-ads/site/internal/browserclient"
)

func TestUnreadMessageActionNavigatesFromMyAds(t *testing.T) {
	page := browserclient.PageAffordances{HasUnreadMessages: true}
	act, ok := unreadMessageAction("/auth/user/myads", page, nil)
	if !ok || act.Action != "get" || act.Path != "/auth/user/messages" {
		t.Fatalf("act %+v ok=%v", act, ok)
	}
}

func TestUnreadMessageActionOpensConversationOnMessagesPage(t *testing.T) {
	page := browserclient.PageAffordances{
		HasUnreadMessages:       true,
		UnreadConversationPaths: []string{"/auth/conversation/5"},
	}
	act, ok := unreadMessageAction("/auth/user/messages", page, nil)
	if !ok || act.Action != "click" || act.Path != "/auth/conversation/5" {
		t.Fatalf("act %+v ok=%v", act, ok)
	}
}

func TestUnreadMessageActionSkipsWhenFormOpen(t *testing.T) {
	page := browserclient.PageAffordances{
		HasUnreadMessages:       true,
		UnreadConversationPaths: []string{"/auth/conversation/5"},
		OpenConversationForms:   []int{5},
	}
	if _, ok := unreadMessageAction("/auth/user/messages", page, nil); ok {
		t.Fatal("expected no unread action when form is open")
	}
}

func TestPendingReplyActionOpensConversation(t *testing.T) {
	inbox := newInbox()
	inbox.update(browserclient.ConversationSnapshot{
		ID: 1,
		Messages: []browserclient.ConversationMessage{
			{Text: "hi", FromSelf: false},
		},
	})
	page := browserclient.PageAffordances{
		HTMX: []browserclient.HTMXAction{
			{Kind: "get", Path: "/auth/conversation/1"},
		},
	}
	act, ok := pendingReplyAction("/auth/user/messages", page, inbox)
	if !ok || act.Action != "click" || act.Path != "/auth/conversation/1" {
		t.Fatalf("act %+v ok=%v", act, ok)
	}
}

func TestPendingReplyActionNavigatesToMessages(t *testing.T) {
	inbox := newInbox()
	inbox.update(browserclient.ConversationSnapshot{
		ID: 1,
		Messages: []browserclient.ConversationMessage{
			{Text: "hi", FromSelf: false},
		},
	})
	act, ok := pendingReplyAction("/auth/user/myads", browserclient.PageAffordances{}, inbox)
	if !ok || act.Action != "get" || act.Path != "/auth/user/messages" {
		t.Fatalf("act %+v ok=%v", act, ok)
	}
}

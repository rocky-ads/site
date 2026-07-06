package browserclient

import (
	"testing"
)

func TestParseConversationHTML(t *testing.T) {
	html := []byte(`<div id="conversation-7-modal">
<div id="conversation-7-messages" class="flex-1 overflow-y-auto p-4 space-y-2">
<div class="flex justify-end mb-2"><div class="max-w-[70%]">
<div class="px-4 py-2 bg-blue-500 text-white rounded-lg rounded-tr-none">Hello there</div>
</div></div>
<div class="flex justify-start mb-2"><div class="max-w-[70%]">
<div class="px-4 py-2 bg-zinc-200 rounded-lg rounded-tl-none">Hi back</div>
</div></div>
</div></div>`)

	conv := ParseConversationHTML(html)
	if conv.ID != 7 || len(conv.Messages) != 2 {
		t.Fatalf("conv %+v", conv)
	}
	if !conv.AwaitingReply() {
		t.Fatal("expected awaiting reply")
	}
}

func TestParsePageUnreadIndicator(t *testing.T) {
	html := []byte(`<div id="message-unread-indicator">
<div class="absolute -top-1 -right-3"><div class="bg-green-500 rounded-full w-3 h-3"></div></div>
</div>`)
	p := ParsePage(html, "/auth/user/myads")
	if !p.HasUnreadMessages {
		t.Fatal("expected unread indicator")
	}
}

func TestParsePageUnreadConversationList(t *testing.T) {
	html := []byte(`<div id="conversation-item-12" hx-get="/auth/conversation/12">
<div class="bg-green-500 rounded-full w-2 h-2"></div>
</div>
<div id="conversation-item-13" hx-get="/auth/conversation/13"></div>`)
	p := ParsePage(html, "/auth/user/messages")
	if len(p.UnreadConversationPaths) != 1 || p.UnreadConversationPaths[0] != "/auth/conversation/12" {
		t.Fatalf("paths %v", p.UnreadConversationPaths)
	}
}

func TestParsePageOpenConversationForm(t *testing.T) {
	html := []byte(`<form id="conversation-3-form" hx-post="/auth/conversation/3/send">
<input name="content" type="text"></form>`)
	p := ParsePage(html, "/auth/user/messages")
	if len(p.OpenConversationForms) != 1 || p.OpenConversationForms[0] != 3 {
		t.Fatalf("forms %v", p.OpenConversationForms)
	}
}

func TestParsePageLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html><html><head><title>Home</title></head>
<body><a href="/login">Login</a><a href="/ad/42">Bike</a></body></html>`)
	p := ParsePage(html, "/")
	if len(p.AdCards) != 1 || p.AdCards[0].ID != "42" {
		t.Fatalf("ad cards %v", p.AdCards)
	}
}

func TestAllowsMessageSendAdPath(t *testing.T) {
	if !AllowsMessageSend("/auth/ad/1/send", nil) {
		t.Fatal("first ad message should be allowed")
	}
	conv := &ConversationSnapshot{ID: 2, Messages: []ConversationMessage{{FromSelf: true}}}
	if AllowsMessageSend("/auth/ad/1/send", conv) {
		t.Fatal("ad send should be blocked when thread exists")
	}
}

func TestFilterMessageSendsRemovesFormWhenNotAwaiting(t *testing.T) {
	conv := ConversationSnapshot{
		ID: 3,
		Messages: []ConversationMessage{
			{Text: "mine", FromSelf: true},
		},
	}
	p := PageAffordances{
		Conversation: &conv,
		Forms: []Form{{
			Action: "/auth/conversation/3/send",
			Method: "POST",
		}},
	}
	out := FilterMessageSends(p)
	if len(out.Forms) != 0 {
		t.Fatalf("forms %v", out.Forms)
	}
}

func TestCheckboxChecked(t *testing.T) {
	if !checkboxChecked("true") || !checkboxChecked("accepted") {
		t.Fatal("expected checked")
	}
	if checkboxChecked("") || checkboxChecked("false") {
		t.Fatal("expected unchecked")
	}
}

func TestIsStuckPage(t *testing.T) {
	if !IsStuckPage(PageAffordances{}) {
		t.Fatal("empty page should be stuck")
	}
	p := PageAffordances{Links: []Link{{Href: "/faq"}}}
	if IsStuckPage(p) {
		t.Fatal("page with links should not be stuck")
	}
}

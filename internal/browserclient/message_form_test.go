package browserclient

import "testing"

func TestAdIDFromSendPath(t *testing.T) {
	id, ok := AdIDFromSendPath("/auth/ad/942/send")
	if !ok || id != 942 {
		t.Fatalf("id=%d ok=%v", id, ok)
	}
}

func TestConversationIDFromOpenPath(t *testing.T) {
	id, ok := ConversationIDFromOpenPath("/auth/conversation/3")
	if !ok || id != 3 {
		t.Fatalf("id=%d ok=%v", id, ok)
	}
	if _, ok := ConversationIDFromOpenPath("/auth/conversation/3/send"); ok {
		t.Fatal("send path should not parse as open path")
	}
}

func TestMessageInputSelectors(t *testing.T) {
	sels := messageInputSelectors("/auth/conversation/2/send")
	if len(sels) != 1 || sels[0] != `#conversation-2-content-input:not([disabled])` {
		t.Fatalf("sels %v", sels)
	}
	adSels := messageInputSelectors("/auth/ad/10/send")
	if len(adSels) != 2 {
		t.Fatalf("ad sels %v", adSels)
	}
}

func TestFilterMessageSendsRemovesSendHTMX(t *testing.T) {
	p := PageAffordances{
		HTMX: []HTMXAction{
			{Kind: "post", Path: "/auth/conversation/1/send"},
			{Kind: "get", Path: "/auth/conversation/1"},
		},
	}
	out := FilterMessageSends(p)
	if len(out.HTMX) != 1 || out.HTMX[0].Path != "/auth/conversation/1" {
		t.Fatalf("htmx %v", out.HTMX)
	}
}

func TestPageHasAdMessageForm(t *testing.T) {
	page := PageAffordances{
		OpenAdMessageSendPaths: []string{"/auth/ad/942/send"},
	}
	if !PageHasAdMessageForm(page, 942) {
		t.Fatal("expected ad message form")
	}
}

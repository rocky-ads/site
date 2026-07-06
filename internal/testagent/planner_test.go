package testagent

import (
	"testing"

	"github.com/rocky-ads/site/internal/browserclient"
)

func TestValidateActionBlocksOwnMessageReply(t *testing.T) {
	conv := &browserclient.ConversationSnapshot{
		ID: 5,
		Messages: []browserclient.ConversationMessage{
			{Text: "hi", FromSelf: true},
		},
	}
	page := browserclient.PageAffordances{
		Conversation: conv,
		Forms: []browserclient.Form{{
			Action: "/auth/conversation/5/send",
			Method: "POST",
		}},
	}
	act := PlannedAction{
		Action: "post_form",
		Path:   "/auth/conversation/5/send",
		Fields: map[string]string{"content": "oops"},
	}
	if err := ValidateAction(act, page, true); err == nil {
		t.Fatal("expected validation error when replying to own message")
	}
}

func TestValidateActionAllowsReplyToOther(t *testing.T) {
	conv := &browserclient.ConversationSnapshot{
		ID: 5,
		Messages: []browserclient.ConversationMessage{
			{Text: "hi", FromSelf: true},
			{Text: "yes", FromSelf: false},
		},
	}
	page := browserclient.EnrichWithConversation(browserclient.PageAffordances{}, *conv)
	act := PlannedAction{
		Action: "post_form",
		Path:   "/auth/conversation/5/send",
		Fields: map[string]string{"content": "thanks"},
	}
	if err := ValidateAction(act, page, true); err != nil {
		t.Fatalf("expected allowed reply: %v", err)
	}
}

func TestValidateActionBlocksAdSendAfterThreadStarted(t *testing.T) {
	conv := &browserclient.ConversationSnapshot{
		ID: 5,
		Messages: []browserclient.ConversationMessage{
			{Text: "hi", FromSelf: true},
		},
	}
	page := browserclient.PageAffordances{
		Conversation: conv,
		Forms: []browserclient.Form{{
			Action: "/auth/ad/370/send",
			Method: "POST",
		}},
	}
	act := PlannedAction{
		Action: "post_form",
		Path:   "/auth/ad/370/send",
		Fields: map[string]string{"content": "again"},
	}
	if err := ValidateAction(act, page, true); err == nil {
		t.Fatal("expected ad send blocked after thread started")
	}
}

func TestValidateActionRejectsClickOnSendPath(t *testing.T) {
	act := PlannedAction{Action: "click", Path: "/auth/ad/942/send"}
	page := browserclient.PageAffordances{
		HTMX: []browserclient.HTMXAction{{Kind: "post", Path: "/auth/ad/942/send"}},
	}
	if err := ValidateAction(act, page, true); err == nil {
		t.Fatal("expected click on send path to be rejected")
	}
}

func TestValidateActionRejectsPasswordForm(t *testing.T) {
	act := PlannedAction{
		Action: "post_form",
		Path:   "/auth/user/settings/password",
		Fields: map[string]string{"current_password": "x"},
	}
	page := browserclient.PageAffordances{
		Forms: []browserclient.Form{{
			Action: "/auth/user/settings/password",
			Method: "POST",
		}},
	}
	if err := ValidateAction(act, page, true); err == nil {
		t.Fatal("expected password form to be rejected")
	}
}

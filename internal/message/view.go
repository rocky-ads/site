package message

import (
	"errors"
	"fmt"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/egg"
	"github.com/rocky-ads/site/internal/user"
)

var (
	ErrModalAdNotFound = errors.New("modal ad not found")
	ErrModalAccess     = errors.New("modal conversation access denied")
)

// ConversationModalView holds data needed to render a conversation modal.
type ConversationModalView struct {
	Conversation     Conversation
	AdTitle          string
	OwnerName        string
	InquirerName     string
	Messages         []Message
	InquirerEggCount int
	OwnerEggCount    int
	CanPost          bool
	HasThrownEgg     bool
	CanThrowEgg      bool
}

func CanViewConversation(conv Conversation, userID int) bool {
	return conv.EggThrowerID != nil ||
		conv.OwnerID == userID ||
		conv.InquirerID == userID
}

func IsParticipant(conv Conversation, userID int) bool {
	return conv.OwnerID == userID || conv.InquirerID == userID
}

// OpenConversation loads a conversation the user may view and marks it read
// when the user is a participant. The second return value is true when marked.
func OpenConversation(conversationID, userID int) (Conversation, bool, error) {
	conv, err := GetConversationByID(conversationID)
	if err != nil {
		return Conversation{}, false, err
	}
	if !CanViewConversation(conv, userID) {
		return Conversation{}, false, ErrModalAccess
	}
	markedRead := false
	if IsParticipant(conv, userID) {
		if err := MarkConversationAsRead(conversationID, userID); err == nil {
			markedRead = true
		}
	}
	return conv, markedRead, nil
}

func OwnerAndInquirerNames(conv Conversation) (ownerName, inquirerName string, err error) {
	owner, err := user.GetByID(conv.OwnerID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get owner name: %w", err)
	}
	inquirer, err := user.GetByID(conv.InquirerID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get inquirer name: %w", err)
	}
	return owner.Name, inquirer.Name, nil
}

func OtherUserName(conv Conversation, currentUserID int) (string, error) {
	var otherUserID int
	if conv.OwnerID == currentUserID {
		otherUserID = conv.InquirerID
	} else {
		otherUserID = conv.OwnerID
	}
	u, err := user.GetByID(otherUserID)
	if err != nil {
		return "", err
	}
	return u.Name, nil
}

func BuildConversationModal(conv Conversation, currentUserID int, loc *time.Location) (ConversationModalView, error) {
	a, err := ad.GetAd(currentUserID, conv.AdID, loc)
	if err != nil {
		return ConversationModalView{}, fmt.Errorf("%w: %v", ErrModalAdNotFound, err)
	}

	ownerName, inquirerName, err := OwnerAndInquirerNames(conv)
	if err != nil {
		return ConversationModalView{}, err
	}

	view := ConversationModalView{
		Conversation:     conv,
		AdTitle:          a.Title,
		OwnerName:        ownerName,
		InquirerName:     inquirerName,
		InquirerEggCount: eggCountForUser(conv.InquirerID),
		OwnerEggCount:    eggCountForUser(conv.OwnerID),
		Messages:         []Message{},
	}

	if conv.ID == 0 {
		view.CanPost = true
		return view, nil
	}

	messages, err := GetConversationMessages(conv.ID, currentUserID, loc)
	if err != nil {
		return ConversationModalView{}, fmt.Errorf("failed to get messages: %w", err)
	}
	view.Messages = messages

	canPost, err := CanUserPost(conv.ID, currentUserID)
	if err != nil {
		return ConversationModalView{}, fmt.Errorf("failed to check permissions: %w", err)
	}
	view.CanPost = canPost

	hasThrown, canThrow := eggThrowPermissions(conv.ID, currentUserID, canPost)
	view.HasThrownEgg = hasThrown
	view.CanThrowEgg = canThrow

	return view, nil
}

func eggCountForUser(userID int) int {
	count, err := egg.GetEggCountForUser(userID)
	if err != nil {
		return 0
	}
	return count
}

func eggThrowPermissions(conversationID, userID int, canPost bool) (hasThrown, canThrow bool) {
	hasThrown, err := egg.HasUserThrownEgg(userID, conversationID)
	if err != nil {
		hasThrown = false
	}
	if !canPost {
		return hasThrown, false
	}
	eggCount, err := egg.GetEggCountForConversation(conversationID)
	if err != nil || eggCount > 0 {
		return hasThrown, false
	}
	userEggCount, err := egg.GetUserEggCount(userID)
	if err != nil || userEggCount >= 3 {
		return hasThrown, false
	}
	return hasThrown, true
}

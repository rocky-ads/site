package testagent

import (
	"sync"

	"github.com/rocky-ads/site/internal/browserclient"
)

type inbox struct {
	mu     sync.Mutex
	byID   map[int]browserclient.ConversationSnapshot
	active int
}

func newInbox() *inbox {
	return &inbox{byID: map[int]browserclient.ConversationSnapshot{}}
}

func (b *inbox) update(conv browserclient.ConversationSnapshot) bool {
	if conv.ID == 0 {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	prev := b.byID[conv.ID]
	b.byID[conv.ID] = conv
	if b.active == 0 {
		b.active = conv.ID
	}
	return conv.ReceivedCount() > prev.ReceivedCount()
}

func (b *inbox) setActive(conv browserclient.ConversationSnapshot) {
	if conv.ID == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.byID[conv.ID] = conv
	b.active = conv.ID
}

func (b *inbox) activeSnapshot() browserclient.ConversationSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.active == 0 {
		return browserclient.ConversationSnapshot{}
	}
	return b.byID[b.active]
}

func (b *inbox) awaitingReplySnapshot() browserclient.ConversationSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, conv := range b.byID {
		if conv.AwaitingReply() {
			return conv
		}
	}
	return browserclient.ConversationSnapshot{}
}

func (b *inbox) enrich(page browserclient.PageAffordances) browserclient.PageAffordances {
	conv := b.awaitingReplySnapshot()
	if conv.ID == 0 {
		conv = b.activeSnapshot()
	}
	if conv.ID == 0 || !browserclient.PageHasOpenConversationForm(page, conv.ID) {
		return page
	}
	return browserclient.EnrichWithConversation(page, conv)
}

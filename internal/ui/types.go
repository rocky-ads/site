package ui

import (
	"time"

	g "maragu.dev/gomponents"
)

// UserProfileData holds presentation fields for a user profile or summary.
type UserProfileData struct {
	Name          string
	MemberSince   string
	ActiveAdCount int
	UserEggCount  int
}

// UserRowData holds presentation fields for an admin user table row.
type UserRowData struct {
	ID        int
	Name      string
	PhoneE64  string
	IsAdmin   bool
	CreatedAt time.Time
	DeletedAt *time.Time
}

// AdHistoryEntry holds one edit-history block on the ad detail page.
type AdHistoryEntry struct {
	Header       string
	Body         string
	ImageIndices []int
}

// AdDetail holds presentation fields for a single ad page.
type AdDetail struct {
	ID                  int
	OwnerID             int
	ImageCount          int
	PriceDisplay        string
	HasPrice            bool
	Title               string
	Location            string
	DescriptionOriginal string
	DescriptionHistory  []AdHistoryEntry
	CreatedAt           time.Time
	Bookmarked          bool
	Active              bool
	Reachable           bool
	RockCount           int
	FacetLabel          string
	FacetDetails        []string // formal facets not shown in title or price row
	Tags                []string
}

// AdCard holds presentation fields for an ad in search or list views.
type AdCard struct {
	ID           int
	PriceDisplay string
	HasPrice     bool
	Title        string
	Location     string
	CreatedAt    time.Time
	ImageCount   int
	Active       bool
	Bookmarked   bool
	RockCount    int
	FacetLabel   string
}

// ConversationModalData holds presentation fields for a conversation modal.
type ConversationModalData struct {
	ConversationID   int
	AdID             int
	OwnerID          int
	InquirerID       int
	CurrentUserID    int
	InquirerEggCount int
	OwnerEggCount    int
	AdTitle          string
	OwnerName        string
	InquirerName     string
	CSRFToken        string
	CanPost          bool
	HasThrownEgg     bool
	CanThrowEgg      bool
	MessageNodes     []g.Node
	EggThrowerID     *int
	TargetModalID    string
}

// ConversationListItemData holds presentation fields for a messages list row.
// LastMessageAt and UpdatedAt are in the viewer's timezone when set.
type ConversationListItemData struct {
	ConversationID     int
	AdID               int
	OwnerID            int
	InquirerID         int
	CurrentUserID      int
	AdTitle            string
	LastMessageContent string
	OtherUserName      string
	LastMessageAt      *time.Time
	UpdatedAt          time.Time
	HasUnread          bool
	EggCount           int
	OtherUserEggCount  int
}

// CategoryOption holds presentation fields for a category picker item.
type CategoryOption struct {
	ID        int
	Name      string
	ImageFile string
}

// SMSQueueEntryInput holds raw queue entry fields for UI formatting.
type SMSQueueEntryInput struct {
	ID            int
	RecipientName string
	AdTitle       string
	Status        string
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}

// MessageItemData holds presentation fields for a single chat message.
// CreatedAt is in the viewer's timezone.
type MessageItemData struct {
	SenderID      int
	CurrentUserID int
	Content       string
	CreatedAt     time.Time
}

// EggEventData holds presentation fields for an egg-thrown timeline entry.
// ThrownAt is in the viewer's timezone.
type EggEventData struct {
	ThrowerID     int
	CurrentUserID int
	ThrownAt      time.Time
	OwnerID       int
	InquirerID    int
}

// QueueStats holds SMS queue counts for the admin dashboard.
type QueueStats struct {
	Pending    int
	Processed  int
	Suppressed int
}

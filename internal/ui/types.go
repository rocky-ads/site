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
	UserRockCount int
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
	IsTest              bool
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
	ConversationID    int
	AdID              int
	OwnerID           int
	InquirerID        int
	CurrentUserID     int
	InquirerRockCount int
	OwnerRockCount    int
	AdTitle           string
	OwnerName         string
	InquirerName      string
	CSRFToken         string
	CanPost           bool
	HasThrownRock     bool
	CanThrowRock      bool
	MessageNodes      []g.Node
	RockThrowerID     *int
	TargetModalID     string
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
	RockCount          int
	OtherUserRockCount int
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

// RockEventData holds presentation fields for a rock-thrown timeline entry.
// ThrownAt is in the viewer's timezone.
type RockEventData struct {
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

// EmbeddingAdminData holds presentation fields for the embeddings admin tab.
type EmbeddingAdminData struct {
	EmbedderProvider string
	EmbedderModel    string
	EmbeddedCount    int
	MissingCount     int
	QueueDepth       int
	Caches           []EmbeddingCachePanel
	MissingAds       []MissingEmbeddingRow
	CategoryID       int
	Categories       []CategoryOption
	UserActivities   []EmbeddingActivityRow
	SiteActivities   []EmbeddingActivityRow
}

// EmbeddingActivityRow holds one weighted activity used to build an embedding.
type EmbeddingActivityRow struct {
	AdID         int
	AdTitle      string
	ActivityType string
	Weight       float32
	Timestamp    string
}

// EmbeddingCachePanel holds cache metrics for one embedding cache tier.
type EmbeddingCachePanel struct {
	Name       string
	Hits       int64
	Misses     int64
	HitRatePct float64
	ItemCount  int64
	MemoryKB   float64
}

// MissingEmbeddingRow holds one ad awaiting vector indexing.
type MissingEmbeddingRow struct {
	AdID         int
	Title        string
	CategoryName string
}

// ClickAdminData holds presentation fields for the clicks admin tab.
type ClickAdminData struct {
	UsersWithClicks int
	AdsClicked      int
	AdDetailViews   int
	ImageNavClicks  int
	ActiveLast7Days int
	TopAds          []ClickAdRow
	TopImages       []ClickImageRow
	RecentActivity  []ClickActivityRow
	TopUsers        []ClickUserRow
}

// ClickAdRow is one row in the top-ads table.
type ClickAdRow struct {
	AdID         int
	Title        string
	CategoryName string
	UserCount    int
	AdViews      int
	ImageClicks  int
	LastActivity string
}

// ClickImageRow is one row in the top-images table.
type ClickImageRow struct {
	AdID       int
	Title      string
	ImageIndex int
	UserCount  int
	Clicks     int
	LastClick  string
}

// ClickActivityRow is one row in recent click activity.
type ClickActivityRow struct {
	When       string
	UserName   string
	UserID     int
	AdID       int
	AdTitle    string
	ClickLabel string
	ClickCount int
}

// ClickUserRow is one row in the top-users table.
type ClickUserRow struct {
	UserID      int
	UserName    string
	AdClicks    int
	ImageClicks int
	LastActive  string
}

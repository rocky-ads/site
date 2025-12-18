package user

import "time"

// Notification method constants
const (
	NotificationMethodSMS    = "sms"
	NotificationMethodEmail  = "email"
	NotificationMethodSignal = "signal"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusArchived UserStatus = "archived"
)

type User struct {
	// Core database fields
	ID        int
	DeletedAt *time.Time
}

// IsArchived returns true if the user has been archived
func (u User) IsArchived() bool {
	return u.DeletedAt != nil
}

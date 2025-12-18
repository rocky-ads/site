package user

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
	ID   int
	Name string
}

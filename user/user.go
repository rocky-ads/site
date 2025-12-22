package user

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rocky-ads/site/db"
)

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
	ID                 int        `json:"id"`
	Name               string     // Decrypted (calculated field)
	EncryptedName      string     `json:"encrypted_name"`
	NameNonce          string     `json:"name_nonce"`
	PasswordHash       string     `json:"password_hash"`
	PasswordSalt       string     `json:"password_salt"`
	PasswordAlgo       string     `json:"password_algo"`
	PhoneE64           string     // Decrypted (calculated field)
	EncryptedPhone     string     `json:"encrypted_phone"`
	PhoneNonce         string     `json:"phone_nonce"`
	EmailAddress       *string    // Decrypted (calculated field)
	EncryptedEmail     *string    `json:"encrypted_email_address"`
	EmailNonce         *string    `json:"email_address_nonce"`
	CreatedAt          time.Time  `json:"created_at"`
	IsAdmin            bool       `json:"is_admin"`
	PhoneVerified      bool       `json:"phone_verified"`
	VerificationCode   *string    `json:"verification_code"`
	NotificationMethod string     `json:"notification_method"`
	SMSOptedOut        bool       `json:"sms_opted_out"`
	DeletedAt          *time.Time `json:"deleted_at"`
}

func GetByID(id int) (u User, err error) {
	query := `SELECT * FROM users WHERE id = ? AND deleted_at IS NULL`
	err = db.QueryRow(query, id).Scan(&u)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func GetByPhoneE64(phoneE64 string) (u User, err error) {
	phoneHash := db.HashString(phoneE64)
	query := `SELECT * FROM users WHERE phone_hash = ? AND deleted_at IS NULL`
	err = db.QueryRow(query, phoneHash).Scan(&u)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func GetByName(name string) (User, error) {
	nameHash := db.HashString(name)

	var u User
	var encryptedNameBytes, nameNonceBytes []byte
	var encryptedPhoneBytes, phoneNonceBytes []byte
	var encryptedEmailBytes, emailNonceBytes []byte
	var phoneVerifiedInt, isAdminInt, smsOptedOutInt int

	query := `SELECT 
		id,
		encrypted_name,
		name_nonce,
		encrypted_phone,
		phone_nonce,
		password_hash,
		password_salt,
		password_algo,
		phone_verified,
		verification_code,
		notification_method,
		encrypted_email,
		email_nonce,
		created_at,
		is_admin,
		sms_opted_out,
		deleted_at
	FROM users WHERE name_hash = ? AND deleted_at IS NULL`

	err := db.QueryRow(query, nameHash).Scan(
		&u.ID,
		&encryptedNameBytes,
		&nameNonceBytes,
		&encryptedPhoneBytes,
		&phoneNonceBytes,
		&u.PasswordHash,
		&u.PasswordSalt,
		&u.PasswordAlgo,
		&phoneVerifiedInt,
		&u.VerificationCode,
		&u.NotificationMethod,
		&encryptedEmailBytes,
		&emailNonceBytes,
		&u.CreatedAt,
		&isAdminInt,
		&smsOptedOutInt,
		&u.DeletedAt,
	)
	if err != nil {
		return User{}, err
	}

	// Convert BLOB to base64 strings
	u.EncryptedName = base64.StdEncoding.EncodeToString(encryptedNameBytes)
	u.NameNonce = base64.StdEncoding.EncodeToString(nameNonceBytes)
	u.EncryptedPhone = base64.StdEncoding.EncodeToString(encryptedPhoneBytes)
	u.PhoneNonce = base64.StdEncoding.EncodeToString(phoneNonceBytes)

	// Convert email BLOBs if present (NULL BLOBs scan as empty slices)
	if len(encryptedEmailBytes) > 0 {
		encryptedEmailStr := base64.StdEncoding.EncodeToString(encryptedEmailBytes)
		u.EncryptedEmail = &encryptedEmailStr
	}
	if len(emailNonceBytes) > 0 {
		emailNonceStr := base64.StdEncoding.EncodeToString(emailNonceBytes)
		u.EmailNonce = &emailNonceStr
	}

	// Convert integer flags to booleans
	u.PhoneVerified = phoneVerifiedInt == 1
	u.IsAdmin = isAdminInt == 1
	u.SMSOptedOut = smsOptedOutInt == 1

	// Decrypt name and phone
	u.Name, err = decryptName(u.ID, u.EncryptedName, u.NameNonce)
	if err != nil {
		return User{}, fmt.Errorf("failed to decrypt name: %w", err)
	}
	u.PhoneE64, err = decryptPhone(u.ID, u.EncryptedPhone, u.PhoneNonce)
	if err != nil {
		return User{}, fmt.Errorf("failed to decrypt phone: %w", err)
	}

	// Decrypt email if present
	if u.EncryptedEmail != nil && u.EmailNonce != nil {
		email, err := decryptEmailAddress(u.ID, *u.EncryptedEmail, *u.EmailNonce)
		if err == nil {
			u.EmailAddress = &email
		}
	}

	return u, nil
}

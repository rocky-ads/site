package user

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/password"
	"github.com/rocky-ads/site/internal/phoneverification"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusArchived UserStatus = "archived"
)

type User struct {
	ID             int
	Name           string // Decrypted (calculated field)
	EncryptedName  string
	NameNonce      string
	PasswordHash   string
	PasswordSalt   string
	PasswordAlgo   string
	PhoneE64       string // Decrypted (calculated field)
	EncryptedPhone string
	PhoneNonce     string
	CreatedAt      time.Time
	IsAdmin        bool
	PhoneVerified  bool
	SMSOptedOut    bool
	DeletedAt      *time.Time
}

const userSelectFields = `SELECT 
	id,
	encrypted_name,
	name_nonce,
	encrypted_phone,
	phone_nonce,
	password_hash,
	password_salt,
	password_algo,
	phone_verified,
	created_at,
	is_admin,
	sms_opted_out,
	deleted_at
FROM users`

func processUserRow(id int, encryptedNameBytes, nameNonceBytes []byte,
	encryptedPhoneBytes, phoneNonceBytes []byte, passwordHash, passwordSalt, passwordAlgo string,
	phoneVerifiedInt int,
	createdAt time.Time, isAdminInt int, smsOptedOutInt int, deletedAt *time.Time,
) (User, error) {
	var u User

	u.ID = id
	u.PasswordHash = passwordHash
	u.PasswordSalt = passwordSalt
	u.PasswordAlgo = passwordAlgo
	u.CreatedAt = createdAt
	u.DeletedAt = deletedAt

	u.EncryptedName = base64.StdEncoding.EncodeToString(encryptedNameBytes)
	u.NameNonce = base64.StdEncoding.EncodeToString(nameNonceBytes)
	u.EncryptedPhone = base64.StdEncoding.EncodeToString(encryptedPhoneBytes)
	u.PhoneNonce = base64.StdEncoding.EncodeToString(phoneNonceBytes)

	u.PhoneVerified = phoneVerifiedInt == 1
	u.IsAdmin = isAdminInt == 1
	u.SMSOptedOut = smsOptedOutInt == 1

	var err error
	u.Name, err = decryptName(u.ID, u.EncryptedName, u.NameNonce)
	if err != nil {
		return User{}, fmt.Errorf("failed to decrypt name: %w", err)
	}
	u.PhoneE64, err = decryptPhone(u.ID, u.EncryptedPhone, u.PhoneNonce)
	if err != nil {
		return User{}, fmt.Errorf("failed to decrypt phone: %w", err)
	}

	return u, nil
}

// scanner interface for both *sql.Row and *sql.Rows
type scanner interface {
	Scan(dest ...interface{}) error
}

func scanUserFields(s scanner) (User, error) {
	var id int
	var encryptedNameBytes, nameNonceBytes []byte
	var encryptedPhoneBytes, phoneNonceBytes []byte
	var phoneVerifiedInt, isAdminInt, smsOptedOutInt int
	var passwordHash, passwordSalt, passwordAlgo string
	var createdAt time.Time
	var deletedAt *time.Time

	err := s.Scan(
		&id,
		&encryptedNameBytes,
		&nameNonceBytes,
		&encryptedPhoneBytes,
		&phoneNonceBytes,
		&passwordHash,
		&passwordSalt,
		&passwordAlgo,
		&phoneVerifiedInt,
		&createdAt,
		&isAdminInt,
		&smsOptedOutInt,
		&deletedAt,
	)
	if err != nil {
		return User{}, err
	}

	return processUserRow(
		id,
		encryptedNameBytes, nameNonceBytes,
		encryptedPhoneBytes, phoneNonceBytes,
		passwordHash, passwordSalt, passwordAlgo,
		phoneVerifiedInt,
		createdAt,
		isAdminInt,
		smsOptedOutInt,
		deletedAt,
	)
}

func getUserBy(whereClause string, args ...any) (User, error) {
	query := userSelectFields + " WHERE " + whereClause
	return scanUserFields(db.QueryRow(query, args...))
}

func GetByID(id int) (User, error) {
	return getUserBy("id = ? AND deleted_at IS NULL", id)
}

// GetByIDIncludingDeleted returns a user by ID, including deleted users
func GetByIDIncludingDeleted(id int) (User, error) {
	return getUserBy("id = ?", id)
}

// Exists checks if a user exists and is not deleted (lightweight check without decryption)
func Exists(id int) bool {
	var exists int
	err := db.QueryRow("SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL", id).Scan(&exists)
	return err == nil
}

func GetByPhoneE64(phoneE64 string) (User, error) {
	phoneHash := db.HashString(phoneE64)
	return getUserBy("phone_hash = ? AND deleted_at IS NULL", phoneHash)
}

func GetByName(name string) (User, error) {
	nameHash := db.HashString(name)
	return getUserBy("name_hash = ? AND deleted_at IS NULL", nameHash)
}

// ErrUserAlreadyExists is returned when attempting to create a user that already exists
var ErrUserAlreadyExists = errors.New("user already exists")

// CreateUser creates a new user with phone verification in a transaction.
// It checks for existing users (including archived), creates the user, marks the phone as verified,
// and cleans up verification codes. Returns the created user or an error.
func CreateUser(username, phoneE64, plainPassword string) (User, error) {

	// Hash password
	passwordHash, passwordSalt, err := password.HashPassword(plainPassword)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return User{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if user already exists (including archived users)
	nameHash := db.HashString(username)
	phoneHash := db.HashString(phoneE64)

	var existingUserID int
	err = tx.QueryRow(`
		SELECT id FROM users 
		WHERE name_hash = $1 OR phone_hash = $2
		LIMIT 1
	`, nameHash, phoneHash).Scan(&existingUserID)
	if err == nil {
		// User exists (including archived)
		return User{}, ErrUserAlreadyExists
	} else if err != sql.ErrNoRows {
		// Database error
		return User{}, fmt.Errorf("failed to check for existing user: %w", err)
	}

	// Insert user with placeholder values to get ID
	result, err := tx.Exec(`
		INSERT INTO users (
			encrypted_name, name_nonce, name_hash,
			password_hash, password_salt, password_algo,
			encrypted_phone, phone_nonce, phone_hash,
			phone_verified, is_admin
		) VALUES ($1, $2, $3, $4, $5, 'argon2id', $6, $7, $8, 0, 0)
	`, []byte{}, []byte{}, nameHash,
		passwordHash, passwordSalt,
		[]byte{}, []byte{}, phoneHash)
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Encrypt name
	encryptedName, nameNonce, err := EncryptName(int(userID), username)
	if err != nil {
		return User{}, fmt.Errorf("failed to encrypt name: %w", err)
	}
	encryptedNameBytes, _ := base64.StdEncoding.DecodeString(encryptedName)
	nameNonceBytes, _ := base64.StdEncoding.DecodeString(nameNonce)

	// Encrypt phone
	encryptedPhone, phoneNonce, err := EncryptPhone(int(userID), phoneE64)
	if err != nil {
		return User{}, fmt.Errorf("failed to encrypt phone: %w", err)
	}
	encryptedPhoneBytes, _ := base64.StdEncoding.DecodeString(encryptedPhone)
	phoneNonceBytes, _ := base64.StdEncoding.DecodeString(phoneNonce)

	// Update user with encrypted fields and mark phone as verified
	_, err = tx.Exec(`
		UPDATE users SET
			encrypted_name = $1, name_nonce = $2,
			encrypted_phone = $3, phone_nonce = $4,
			phone_verified = 1
		WHERE id = $5
	`, encryptedNameBytes, nameNonceBytes,
		encryptedPhoneBytes, phoneNonceBytes,
		userID)
	if err != nil {
		return User{}, fmt.Errorf("failed to update user with encrypted fields: %w", err)
	}

	// Cleanup registration validation codes
	if err := phoneverification.InvalidateCodesTx(tx, phoneE64); err != nil {
		return User{}, fmt.Errorf("failed to cleanup verification codes: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get the created user
	u, err := GetByID(int(userID))
	if err != nil {
		return User{}, fmt.Errorf("failed to get created user: %w", err)
	}

	return u, nil
}

func SetSMSOptOut(userID int, optedOut bool) error {
	val := 0
	if optedOut {
		val = 1
	}
	_, err := db.Exec("UPDATE Users SET sms_opted_out = $1 WHERE id = $2", val, userID)
	return err
}

// GetAllUsers returns all users (including deleted) with optional sorting
func GetAllUsers(sortBy string, sortOrder string) ([]User, error) {
	validSortColumns := map[string]bool{
		"id":         true,
		"name":       true,
		"phone":      true,
		"is_admin":   true,
		"created_at": true,
		"deleted_at": true,
	}

	if !validSortColumns[sortBy] {
		sortBy = "id"
	}

	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}

	query := userSelectFields

	needsInMemorySort := sortBy == "name" || sortBy == "phone"
	if !needsInMemorySort {
		query += " ORDER BY " + sortBy + " " + sortOrder
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		u, err := scanUserFields(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if needsInMemorySort {
		sort.Slice(users, func(i, j int) bool {
			var less bool
			switch sortBy {
			case "name":
				less = users[i].Name < users[j].Name
			case "phone":
				less = users[i].PhoneE64 < users[j].PhoneE64
			}
			if sortOrder == "DESC" {
				return !less
			}
			return less
		})
	}

	return users, nil
}

func DeleteUser(userID int) error {
	now := time.Now()
	_, err := db.Exec("UPDATE users SET deleted_at = ? WHERE id = ?", now, userID)
	return err
}

func RestoreUser(userID int) error {
	_, err := db.Exec("UPDATE users SET deleted_at = NULL WHERE id = ?", userID)
	return err
}

func PromoteToAdmin(userID int) error {
	_, err := db.Exec("UPDATE users SET is_admin = 1 WHERE id = ?", userID)
	return err
}

func DemoteFromAdmin(userID int) error {
	_, err := db.Exec("UPDATE users SET is_admin = 0 WHERE id = ?", userID)
	return err
}

// IsReachable reports whether the user can be contacted via the platform.
// Users must verify their phone during registration before they can receive messages.
// TODO: incorporate notification preferences and SMS opt-out.
func IsReachable(userID int) (bool, error) {
	u, err := GetByID(userID)
	if err != nil {
		return false, err
	}
	return u.PhoneVerified, nil
}

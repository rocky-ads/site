package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/encryption"
)

func resolveUserID(u UserRow) (int, error) {
	var id int
	err := db.QueryRow(
		`SELECT id FROM users WHERE name_hash = $1`, u.NameHash,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("lookup user %s: %w", u.NameHash, err)
	}
	if len(config.UserEncryptionKey) == 0 {
		return 0, fmt.Errorf(
			"USER_ENCRYPTION_KEY required to restore new user %s",
			u.NameHash,
		)
	}

	err = db.QueryRow(`
		INSERT INTO users (
			created_at, deleted_at, is_admin,
			encrypted_name, name_nonce, name_hash,
			password_hash, password_salt, password_algo,
			encrypted_phone, phone_nonce, phone_hash,
			phone_verified, sms_opted_out, last_sms_sent_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			$7, $8, $9,
			$10, $11, $12,
			$13, $14, $15
		) RETURNING id`,
		u.CreatedAt, u.DeletedAt, u.IsAdmin,
		[]byte{}, []byte{}, u.NameHash,
		u.PasswordHash, u.PasswordSalt, u.PasswordAlgo,
		[]byte{}, []byte{}, u.PhoneHash,
		u.PhoneVerified, u.SMSOptedOut, u.LastSMSSentAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert user %s: %w", u.NameHash, err)
	}

	decryptKey, err := userDecryptKey()
	if err != nil {
		return 0, err
	}

	name, err := decryptField(
		decryptKey, u.EncryptUserID, []byte(u.EncryptedName), []byte(u.NameNonce),
	)
	if err != nil {
		return 0, fmt.Errorf("decrypt name for user %s: %w", u.NameHash, err)
	}
	phone, err := decryptField(
		decryptKey, u.EncryptUserID, []byte(u.EncryptedPhone), []byte(u.PhoneNonce),
	)
	if err != nil {
		return 0, fmt.Errorf("decrypt phone for user %s: %w", u.NameHash, err)
	}

	encName, nameNonce, err := encryption.Encrypt(
		id, name, config.UserEncryptionKey,
	)
	if err != nil {
		return 0, fmt.Errorf("encrypt name for user %s: %w", u.NameHash, err)
	}
	encPhone, phoneNonce, err := encryption.Encrypt(
		id, phone, config.UserEncryptionKey,
	)
	if err != nil {
		return 0, fmt.Errorf("encrypt phone for user %s: %w", u.NameHash, err)
	}

	encNameBytes, _ := base64.StdEncoding.DecodeString(encName)
	nameNonceBytes, _ := base64.StdEncoding.DecodeString(nameNonce)
	encPhoneBytes, _ := base64.StdEncoding.DecodeString(encPhone)
	phoneNonceBytes, _ := base64.StdEncoding.DecodeString(phoneNonce)

	_, err = db.Exec(`
		UPDATE users SET
			encrypted_name = $1, name_nonce = $2,
			encrypted_phone = $3, phone_nonce = $4
		WHERE id = $5`,
		encNameBytes, nameNonceBytes,
		encPhoneBytes, phoneNonceBytes,
		id,
	)
	if err != nil {
		return 0, fmt.Errorf("update encrypted user %s: %w", u.NameHash, err)
	}
	return id, nil
}

func userDecryptKey() ([]byte, error) {
	if key := encryptionKeyFromEnv("BACKUP_USER_ENCRYPTION_KEY"); len(key) == 32 {
		return key, nil
	}
	if len(config.UserEncryptionKey) == 32 {
		return config.UserEncryptionKey, nil
	}
	return nil, fmt.Errorf(
		"BACKUP_USER_ENCRYPTION_KEY or USER_ENCRYPTION_KEY required to decrypt users",
	)
}

func encryptionKeyFromEnv(name string) []byte {
	keyStr := os.Getenv(name)
	if keyStr == "" {
		return nil
	}
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil
	}
	return key
}

func decryptField(key []byte, encryptID int, ciphertext, nonce []byte) (string, error) {
	return encryption.Decrypt(
		encryptID,
		base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(nonce),
		key,
	)
}

package user

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/encryption"
)

func EncryptName(userID int, name string) (encryptedName, nonce string, err error) {
	return encryption.Encrypt(userID, name, config.UserEncryptionKey)
}

func decryptName(userID int, encryptedName, nonce string) (name string, err error) {
	return encryption.Decrypt(userID, encryptedName, nonce, config.UserEncryptionKey)
}

func EncryptPhone(userID int, phone string) (encryptedPhone, nonce string, err error) {
	return encryption.Encrypt(userID, phone, config.UserEncryptionKey)
}

func decryptPhone(userID int, encryptedPhone, nonce string) (phone string, err error) {
	return encryption.Decrypt(userID, encryptedPhone, nonce, config.UserEncryptionKey)
}

func EncryptEmailAddress(userID int, emailAddress string) (encryptedEmail, nonce string, err error) {
	return encryption.Encrypt(userID, emailAddress, config.UserEncryptionKey)
}

func decryptEmailAddress(userID int, encryptedEmail, nonce string) (emailAddress string, err error) {
	return encryption.Decrypt(userID, encryptedEmail, nonce, config.UserEncryptionKey)
}

// hashString creates a SHA256 hash of a value for database lookups
// This allows us to search for users by phone/name without decrypting all records
func hashString(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

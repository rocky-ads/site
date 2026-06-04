package user

import (
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/encryption"
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

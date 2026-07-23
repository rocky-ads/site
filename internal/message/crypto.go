package message

import (
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/encryption"
)

func openJournal(id int, stored string) (string, error) {
	return encryption.Open(id, stored, config.DBEncryptionKey)
}

func sealJournal(id int, plain string) (string, error) {
	return encryption.Seal(id, plain, config.DBEncryptionKey)
}

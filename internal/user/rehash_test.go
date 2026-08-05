package user

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/rocky-ads/site/internal/db"
)

func TestRehashLookupHashesUpgradesUnsalted(t *testing.T) {
	resetSchema(t)

	u, err := CreateUser("rehashuser", "+15550109901", "password1")
	if err != nil {
		t.Fatal(err)
	}

	plainName := sha256.Sum256([]byte(u.Name))
	plainPhone := sha256.Sum256([]byte(u.PhoneE64))
	_, err = db.Exec(`
		UPDATE users SET name_hash = $1, phone_hash = $2 WHERE id = $3`,
		hex.EncodeToString(plainName[:]),
		hex.EncodeToString(plainPhone[:]),
		u.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	n, err := RehashLookupHashes()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("updated=%d want 1", n)
	}

	got, err := GetByName("rehashuser")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != u.ID {
		t.Fatalf("lookup after rehash failed")
	}

	n, err = RehashLookupHashes()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("second pass updated=%d want 0", n)
	}
}

package phoneverification

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/db/testdb"
	"github.com/rocky-ads/site/internal/logger"
)

func TestMain(m *testing.M) {
	if err := chdirModuleRoot(); err != nil {
		panic(err)
	}
	testURL := testdb.PackageDatabaseURL("phoneverification")
	if err := testdb.EnsureDatabase(testURL); err != nil {
		panic(err)
	}
	os.Setenv("DATABASE_URL", testURL)
	if err := logger.Init("error", "text", ""); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func chdirModuleRoot() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return os.Chdir(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func resetSchema(t *testing.T) {
	t.Helper()
	if err := testdb.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}
}

func TestPurposeIsolation(t *testing.T) {
	resetSchema(t)

	const phone = "+15559872001"
	registerCode := "111111"
	changeCode := "222222"
	userID := 42

	if err := insertStubUser(userID); err != nil {
		t.Fatalf("stub user: %v", err)
	}

	if err := StoreCode(phone, registerCode, PurposeRegister, nil); err != nil {
		t.Fatalf("store register: %v", err)
	}
	if err := StoreCode(phone, changeCode, PurposeChangePhone, &userID); err != nil {
		t.Fatalf("store change_phone: %v", err)
	}

	ok, err := ValidateCode(phone, registerCode, PurposeChangePhone, &userID)
	if err == nil && ok {
		t.Fatal("register code must not validate as change_phone")
	}

	ok, err = ValidateCode(phone, changeCode, PurposeRegister, nil)
	if err == nil && ok {
		t.Fatal("change_phone code must not validate as register")
	}

	ok, err = ValidateCode(phone, registerCode, PurposeRegister, nil)
	if err != nil || !ok {
		t.Fatalf("register code valid: ok=%v err=%v", ok, err)
	}

	ok, err = ValidateCode(phone, changeCode, PurposeChangePhone, &userID)
	if err != nil || !ok {
		t.Fatalf("change_phone code valid: ok=%v err=%v", ok, err)
	}
}

func TestInvalidateCodesForPurpose(t *testing.T) {
	resetSchema(t)

	const phone = "+15559872002"
	userID := 7
	if err := insertStubUser(userID); err != nil {
		t.Fatalf("stub user: %v", err)
	}

	if err := StoreCode(phone, "333333", PurposeRegister, nil); err != nil {
		t.Fatalf("store register: %v", err)
	}
	if err := StoreCode(phone, "444444", PurposeChangePhone, &userID); err != nil {
		t.Fatalf("store change: %v", err)
	}

	if err := InvalidateCodesForPurpose(phone, PurposeRegister, nil); err != nil {
		t.Fatalf("invalidate register: %v", err)
	}

	ok, err := ValidateCode(phone, "333333", PurposeRegister, nil)
	if err == nil && ok {
		t.Fatal("register code should be gone")
	}

	ok, err = ValidateCode(phone, "444444", PurposeChangePhone, &userID)
	if err != nil || !ok {
		t.Fatalf("change_phone code should remain: ok=%v err=%v", ok, err)
	}
}

func insertStubUser(userID int) error {
	hash := fmt.Sprintf("stub-hash-%d", userID)
	_, err := db.Exec(`
		INSERT INTO users (
			id, encrypted_name, name_nonce, name_hash,
			password_hash, password_salt, password_algo,
			encrypted_phone, phone_nonce, phone_hash,
			phone_verified, is_admin
		) VALUES (
			$1, '\x00', '\x00', $2,
			'h', 's', 'argon2id',
			'\x00', '\x00', $3,
			1, 0
		)
	`, userID, hash+"-name", hash+"-phone")
	return err
}

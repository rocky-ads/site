package main

import (
	"fmt"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
)

const phoneHashActiveIndex = "idx_users_phone_hash_active"
const phoneVerificationPurposeCheck = "phone_verification_purpose_check"
const phoneVerificationPurposeIndex = "idx_phone_verification_purpose"

// migratePhoneLifecycleSchema upgrades an old DB to the phone/username
// lifecycle schema: phone unique among live users only, and purpose-bound
// phone_verification. Idempotent — safe to run on an already-new DB.
func migratePhoneLifecycleSchema(dryRun bool) error {
	needsPhone, err := needsPhoneHashPartialUnique()
	if err != nil {
		return err
	}
	needsPV, err := needsPhoneVerificationPurpose()
	if err != nil {
		return err
	}
	if !needsPhone && !needsPV {
		logger.Info("Phone lifecycle schema already up to date")
		return nil
	}

	if dryRun {
		logger.Info("Dry run migrate-schema",
			"phone_hash_partial_unique", needsPhone,
			"phone_verification_purpose", needsPV)
		return nil
	}

	if needsPhone {
		if err := migratePhoneHashUnique(); err != nil {
			return err
		}
	}
	if needsPV {
		if err := migratePhoneVerificationPurpose(); err != nil {
			return err
		}
	}

	logger.Info("Phone lifecycle schema migration complete")
	return nil
}

func needsPhoneHashPartialUnique() (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE schemaname = 'public'
			  AND indexname = $1
		)
	`, phoneHashActiveIndex).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check %s: %w", phoneHashActiveIndex, err)
	}
	return !exists, nil
}

func needsPhoneVerificationPurpose() (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'phone_verification'
			  AND column_name = 'purpose'
		)
	`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check phone_verification.purpose: %w", err)
	}
	return !exists, nil
}

func migratePhoneHashUnique() error {
	// Drop column-level UNIQUE on phone_hash (name is typically
	// users_phone_hash_key when defined as UNIQUE in CREATE TABLE).
	rows, err := db.Query(`
		SELECT c.conname
		FROM pg_constraint c
		JOIN pg_class t ON c.conrelid = t.oid
		JOIN pg_namespace n ON t.relnamespace = n.oid
		WHERE n.nspname = 'public'
		  AND t.relname = 'users'
		  AND c.contype = 'u'
		  AND pg_get_constraintdef(c.oid) LIKE '%phone_hash%'
	`)
	if err != nil {
		return fmt.Errorf("list phone_hash unique constraints: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan constraint name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, name := range names {
		logger.Info("Dropping users phone_hash unique constraint",
			"constraint", name)
		_, err := db.Exec(fmt.Sprintf(
			`ALTER TABLE users DROP CONSTRAINT %s`,
			quoteIdent(name),
		))
		if err != nil {
			return fmt.Errorf("drop constraint %s: %w", name, err)
		}
	}

	logger.Info("Creating partial unique index on live phone_hash")
	_, err = db.Exec(fmt.Sprintf(`
		CREATE UNIQUE INDEX %s
		ON users(phone_hash) WHERE deleted_at IS NULL
	`, phoneHashActiveIndex))
	if err != nil {
		return fmt.Errorf("create %s: %w", phoneHashActiveIndex, err)
	}
	return nil
}

func migratePhoneVerificationPurpose() error {
	logger.Info("Adding phone_verification purpose and user_id columns")

	// Drop ephemeral codes; old rows have no purpose and expire quickly.
	if _, err := db.Exec(`DELETE FROM phone_verification`); err != nil {
		return fmt.Errorf("clear phone_verification: %w", err)
	}

	_, err := db.Exec(`
		ALTER TABLE phone_verification
		ADD COLUMN purpose TEXT,
		ADD COLUMN user_id INTEGER REFERENCES users(id)
	`)
	if err != nil {
		return fmt.Errorf("add phone_verification columns: %w", err)
	}

	_, err = db.Exec(`
		ALTER TABLE phone_verification
		ALTER COLUMN purpose SET NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("set purpose NOT NULL: %w", err)
	}

	_, err = db.Exec(fmt.Sprintf(`
		ALTER TABLE phone_verification
		ADD CONSTRAINT %s CHECK (
			(purpose = 'register' AND user_id IS NULL) OR
			(purpose = 'change_phone' AND user_id IS NOT NULL)
		)
	`, phoneVerificationPurposeCheck))
	if err != nil {
		return fmt.Errorf("add purpose check: %w", err)
	}

	_, err = db.Exec(fmt.Sprintf(`
		CREATE INDEX %s
		ON phone_verification(phone_e64, purpose, user_id)
	`, phoneVerificationPurposeIndex))
	if err != nil {
		return fmt.Errorf("create purpose index: %w", err)
	}
	return nil
}

func quoteIdent(name string) string {
	return `"` + name + `"`
}

package eggopinion

import (
	"os"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/message"
)

func setupEggOpinionDB(t *testing.T) {
	t.Helper()
	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(":memory:"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO users (encrypted_name, name_nonce, name_hash, password_hash, password_salt, encrypted_phone, phone_nonce, phone_hash)
		VALUES (x'', x'', 'owner', 'p', 's', x'', x'', 'ph'),
		       (x'', x'', 'enquirer', 'p', 's', x'', x'', 'ph2')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO categories (name, seed_ad_file, image_file)
		VALUES ('Test', 'a.json', 'c.svg')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO ads (category_id, title, description, user_id)
		VALUES (1, 'Test ad', 'desc', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO conversations (ad_id, owner_id, enquirer_id, egg_thrower_id, egg_thrown_at)
		VALUES (1, 1, 2, 2, CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatal(err)
	}
}

func insertTestOpinion(t *testing.T) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO egg_opinions (
		conversation_id, summary, assessment, assessment_detail,
		resolution, reasoning
	) VALUES (1, 'Cached summary text.', 5,
		'Both sides acted reasonably.', 'Keep talking.',
		'No policy violation.')`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEggOpinionDB(t *testing.T) {
	setupEggOpinionDB(t)

	conv, err := message.GetConversationByID(1)
	if err != nil {
		t.Fatal(err)
	}
	loc := time.UTC

	t.Run("returns cached opinion", func(t *testing.T) {
		insertTestOpinion(t)

		op, err := GetOrGenerate(conv, loc)
		if err != nil {
			t.Fatal(err)
		}
		if op.Summary != "Cached summary text." {
			t.Fatalf("summary = %q", op.Summary)
		}
		if op.Assessment != 5 {
			t.Fatalf("assessment = %d", op.Assessment)
		}
	})

	t.Run("invalidate conversation", func(t *testing.T) {
		if err := Invalidate(1); err != nil {
			t.Fatal(err)
		}
		var count int
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM egg_opinions WHERE conversation_id = 1`,
		).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatal("expected cache cleared")
		}
	})

	t.Run("invalidate for ad", func(t *testing.T) {
		insertTestOpinion(t)

		if err := InvalidateForAd(1); err != nil {
			t.Fatal(err)
		}
		var count int
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM egg_opinions WHERE conversation_id = 1`,
		).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatal("expected cache cleared for ad")
		}
	})
}

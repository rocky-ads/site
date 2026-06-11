package message

import (
	"os"
	"testing"

	"github.com/rocky-ads/site/internal/db"
)

func TestOpenConversationReadMarking(t *testing.T) {
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
		       (x'', x'', 'inquirer', 'p', 's', x'', x'', 'ph2'),
		       (x'', x'', 'viewer', 'p', 's', x'', x'', 'ph3')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO categories (name, seed_ad_file, image_file) VALUES ('Test', 'a.json', 'c.svg')`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO ads (category_id, title, description, user_id)
		VALUES (1, 'Test ad', 'desc', 1)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO conversations (ad_id, owner_id, inquirer_id, inquirer_has_unread, egg_thrower_id, egg_thrown_at)
		VALUES (1, 1, 2, 1, 2, CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("skips read for non-participant", func(t *testing.T) {
		_, marked, err := OpenConversation(1, 3)
		if err != nil {
			t.Fatal(err)
		}
		if marked {
			t.Fatal("non-participant should not mark conversation read")
		}
		var unread int
		if err := db.QueryRow(
			`SELECT inquirer_has_unread FROM conversations WHERE id = 1`,
		).Scan(&unread); err != nil {
			t.Fatal(err)
		}
		if unread != 1 {
			t.Fatalf("inquirer_has_unread = %d, want 1", unread)
		}
	})

	t.Run("marks read for participant", func(t *testing.T) {
		_, err := db.Exec(
			`UPDATE conversations SET inquirer_has_unread = 1 WHERE id = 1`,
		)
		if err != nil {
			t.Fatal(err)
		}
		_, marked, err := OpenConversation(1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if !marked {
			t.Fatal("participant should mark conversation read")
		}
		var unread int
		if err := db.QueryRow(
			`SELECT inquirer_has_unread FROM conversations WHERE id = 1`,
		).Scan(&unread); err != nil {
			t.Fatal(err)
		}
		if unread != 0 {
			t.Fatalf("inquirer_has_unread = %d, want 0", unread)
		}
	})
}

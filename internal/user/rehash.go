package user

import (
	"encoding/base64"
	"fmt"

	"github.com/rocky-ads/site/internal/db"
)

// RehashLookupHashes rewrites name_hash/phone_hash from decrypted
// plaintext using the current peppered HashString. Idempotent.
func RehashLookupHashes() (updated int, err error) {
	rows, err := db.Query(`
		SELECT id, encrypted_name, name_nonce, encrypted_phone, phone_nonce,
			name_hash, phone_hash
		FROM users`)
	if err != nil {
		return 0, fmt.Errorf("rehash lookup hashes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var encName, nameNonce, encPhone, phoneNonce []byte
		var nameHash, phoneHash string
		if err := rows.Scan(
			&id, &encName, &nameNonce, &encPhone, &phoneNonce,
			&nameHash, &phoneHash,
		); err != nil {
			return updated, fmt.Errorf("rehash scan: %w", err)
		}
		if len(encName) == 0 || len(encPhone) == 0 {
			continue
		}

		name, err := decryptName(id,
			base64.StdEncoding.EncodeToString(encName),
			base64.StdEncoding.EncodeToString(nameNonce))
		if err != nil {
			return updated, fmt.Errorf("rehash decrypt name id=%d: %w", id, err)
		}
		phone, err := decryptPhone(id,
			base64.StdEncoding.EncodeToString(encPhone),
			base64.StdEncoding.EncodeToString(phoneNonce))
		if err != nil {
			return updated, fmt.Errorf("rehash decrypt phone id=%d: %w", id, err)
		}

		wantName := db.HashString(name)
		wantPhone := db.HashString(phone)
		if wantName == nameHash && wantPhone == phoneHash {
			continue
		}

		_, err = db.Exec(`
			UPDATE users SET name_hash = $1, phone_hash = $2
			WHERE id = $3`,
			wantName, wantPhone, id,
		)
		if err != nil {
			return updated, fmt.Errorf("rehash update id=%d: %w", id, err)
		}
		updated++
	}
	if err := rows.Err(); err != nil {
		return updated, fmt.Errorf("rehash rows: %w", err)
	}
	return updated, nil
}

package db

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHashStringRequiresPepper(t *testing.T) {
	prev := hashPepper
	hashPepper = nil
	t.Cleanup(func() { hashPepper = prev })
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic without pepper")
		}
	}()
	_ = HashString("x")
}

func TestHashStringPeppered(t *testing.T) {
	pepper := bytes32(1)
	SetHashPepper(pepper)
	t.Cleanup(func() { hashPepper = nil })

	a := HashString("+15551234567")
	b := HashString("+15551234567")
	if a != b {
		t.Fatal("same input must hash equal")
	}
	if a == HashString("+15557654321") {
		t.Fatal("different inputs must differ")
	}

	plain := sha256.Sum256([]byte("+15551234567"))
	if a == hex.EncodeToString(plain[:]) {
		t.Fatal("peppered hash must not equal unsalted SHA-256")
	}

	SetHashPepper(bytes32(2))
	if HashString("+15551234567") == a {
		t.Fatal("different pepper must change hash")
	}
}

func bytes32(b byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = b
	}
	return out
}

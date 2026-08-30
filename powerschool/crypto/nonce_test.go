package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"
)

// Nonce() is intentionally random — assert shape (byte length, valid
// hex/base64), not a fixed value.

func TestNonce_HexNoBase64(t *testing.T) {
	nonce, date := Nonce(8, false)
	raw, err := hex.DecodeString(nonce)
	if err != nil {
		t.Fatalf("not valid hex: %v", err)
	}
	if len(raw) != 8 {
		t.Fatalf("want 8 raw bytes, got %d", len(raw))
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000Z", date); err != nil {
		t.Fatalf("date not in expected format: %v", err)
	}
}

func TestNonce_HexBase64(t *testing.T) {
	nonce, _ := Nonce(16, true)
	decoded, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		t.Fatalf("not valid base64: %v", err)
	}
	// decoded payload is the ASCII hex encoding of 16 random bytes = 32 chars
	if len(decoded) != 32 {
		t.Fatalf("want 32 decoded bytes (hex of 16 raw bytes), got %d", len(decoded))
	}
	if _, err := hex.DecodeString(string(decoded)); err != nil {
		t.Fatalf("base64-decoded payload isn't valid hex: %v", err)
	}
}

func TestNonce_Uniqueness(t *testing.T) {
	a, _ := Nonce(8, false)
	b, _ := Nonce(8, false)
	if a == b {
		t.Fatal("two calls produced the same nonce — rand.Reader broken or not being used")
	}
}

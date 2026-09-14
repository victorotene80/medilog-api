package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type SessionKeyHasher struct {
	pepper string
}

func NewSessionKeyHasher(pepper string) (*SessionKeyHasher, error) {
	if pepper == "" {
		return nil, errors.New("session pepper cannot be empty")
	}

	return &SessionKeyHasher{
		pepper: pepper,
	}, nil
}

// Hash derives the value stored in refresh_tokens.token_hash and referenced by
// replaced_by_token_hash.
//
// This is a rotation breadcrumb, not an authentication credential: the raw key
// is never returned to any client, so nothing can present it and there is
// nothing to verify it against. Making it a real second factor would mean
// handing the raw key to the client and looking the row up by hash on refresh —
// a change to what the refresh token *is*, not a bug fix. Until that is a
// deliberate decision, do not read this as a security control.
func (h *SessionKeyHasher) Hash(rawKey string) string {
	data := h.pepper + rawKey
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("invalid length for random string")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

package services

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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

func (h *SessionKeyHasher) Hash(rawKey string) string {
	data := h.pepper + rawKey
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (h *SessionKeyHasher) Verify(rawKey, storedHash string) bool {
	computed := h.Hash(rawKey)

	if len(computed) != len(storedHash) {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(computed),
		[]byte(storedHash),
	) == 1
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

package valueobjects

import (
	"errors"
	"strings"
)

type PasswordHash string

func NewPasswordHash(value string) (PasswordHash, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("password hash is required")
	}
	return PasswordHash(value), nil
}

func (p PasswordHash) String() string { return string(p) }

func PasswordHashFromString(s string) PasswordHash {
	return PasswordHash(s)
}

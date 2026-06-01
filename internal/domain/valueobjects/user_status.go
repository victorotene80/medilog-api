package valueobjects

import (
	"errors"
)

type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusDeleted             UserStatus = "deleted"
	UserStatusLocked              UserStatus = "locked"
	UserStatusPendingVerification UserStatus = "pending_verification"
)

func NewUserStatus(raw string) (UserStatus, error) {
	s := UserStatus(raw)
	switch s {
	case UserStatusActive, UserStatusSuspended, UserStatusDeleted, UserStatusLocked, UserStatusPendingVerification:
		return s, nil
	}
	return "", errors.New("invalid user status")
}

func (s UserStatus) String() string              { return string(s) }
func (s UserStatus) IsActive() bool              { return s == UserStatusActive }
func (s UserStatus) IsDeleted() bool             { return s == UserStatusDeleted }
func (s UserStatus) IsLocked() bool              { return s == UserStatusLocked }
func (s UserStatus) IsPendingVerification() bool { return s == UserStatusPendingVerification }

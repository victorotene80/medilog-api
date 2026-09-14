package valueobjects

import "errors"

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

func NewUserRole(raw string) (UserRole, error) {
	r := UserRole(raw)
	switch r {
	case UserRoleUser, UserRoleAdmin:
		return r, nil
	}
	return "", errors.New("invalid user role")
}

func (r UserRole) String() string { return string(r) }

func (r UserRole) IsAdmin() bool { return r == UserRoleAdmin }

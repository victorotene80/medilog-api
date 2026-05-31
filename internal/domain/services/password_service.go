package services

import (
	"unicode"

	"github.com/victorotene80/medilog-api/internal/domain"
	"github.com/victorotene80/medilog-api/internal/domain/services/policy"
)

type PasswordService struct {
	policy policy.PasswordPolicy
}

func NewPasswordService(policy policy.PasswordPolicy) *PasswordService {
	return &PasswordService{policy: policy}
}

func (s *PasswordService) Validate(password string) error {
	if len(password) < s.policy.MinLength {
		return domain.ErrPasswordTooShort
	}
	if s.policy.RequireUppercase && !containsUppercase(password) {
		return domain.ErrPasswordMissingUppercase
	}
	if s.policy.RequireLowercase && !containsLowercase(password) {
		return domain.ErrPasswordMissingLowercase
	}
	if s.policy.RequireNumbers && !containsNumber(password) {
		return domain.ErrPasswordMissingNumber
	}
	if s.policy.RequireSpecialChar && !containsSpecialChar(password) {
		return domain.ErrPasswordMissingSpecial
	}
	return nil
}

func containsUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func containsLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func containsSpecialChar(s string) bool {
	special := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, r := range s {
		for _, sc := range special {
			if r == sc {
				return true
			}
		}
	}
	return false
}

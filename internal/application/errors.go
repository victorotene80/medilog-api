package application

import "errors"

var (
	ErrHandlerExists      = errors.New("handler already registered for command")
	ErrHandlerNotFound    = errors.New("handler not found for command")
	ErrNilCommand         = errors.New("command cannot be nil")
	ErrInvalidResult      = errors.New("invalid result type")
	ErrSessionInvalid     = errors.New("session invalid or expired")
	ErrCacheMiss          = errors.New("cache miss")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrUserNotFound       = errors.New("user not found")
	ErrAccountLocked      = errors.New("account locked")
	ErrVerificationNeeded = errors.New("verification required")
	ErrOnboardingRequired = errors.New("onboarding required")
)

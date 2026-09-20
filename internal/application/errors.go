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

// ErrorKind classifies an AppError so the HTTP layer can map it to a
// meaningful status code instead of collapsing everything into a generic 500.
type ErrorKind int

const (
	KindInternal ErrorKind = iota
	KindNotFound
	KindConflict
	KindValidation
	KindUnauthorized
	KindForbidden
	KindRateLimited
	// KindQuotaExceeded means the caller's allowance for a metered resource is
	// spent (e.g. the daily AI question quota). It is deliberately distinct from
	// KindRateLimited: rate limiting means "too fast, retry shortly", whereas a
	// spent quota does not clear until the window resets or the plan is upgraded.
	// The HTTP layer maps it to 402 so clients can tell the two apart.
	KindQuotaExceeded
	// KindOffTopic means the AI assistant rejected a question as unrelated to
	// health/medication. It is deliberately distinct from KindValidation so
	// clients can show a specific "ask something health-related" message
	// instead of a generic bad-input message.
	KindOffTopic
)

// AppError is a typed application error carrying a machine-friendly kind.
// It flows through the command bus and httperr.StatusFrom (via errors.As),
// so domain failures surface with a correct HTTP status.
type AppError struct {
	Kind    ErrorKind
	Message string
}

func (e *AppError) Error() string { return e.Message }

func NewNotFound(msg string) error {
	return &AppError{Kind: KindNotFound, Message: msg}
}

func NewConflict(msg string) error {
	return &AppError{Kind: KindConflict, Message: msg}
}

func NewValidation(msg string) error {
	return &AppError{Kind: KindValidation, Message: msg}
}

func NewUnauthorized(msg string) error {
	return &AppError{Kind: KindUnauthorized, Message: msg}
}

func NewForbidden(msg string) error {
	return &AppError{Kind: KindForbidden, Message: msg}
}

func NewRateLimited(msg string) error {
	return &AppError{Kind: KindRateLimited, Message: msg}
}

func NewQuotaExceeded(msg string) error {
	return &AppError{Kind: KindQuotaExceeded, Message: msg}
}

func NewOffTopic(msg string) error {
	return &AppError{Kind: KindOffTopic, Message: msg}
}

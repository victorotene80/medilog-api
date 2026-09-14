package contracts

import (
	"context"
	"time"
)

// SessionInvalidator ends every live session belonging to a user.
//
// It exists because ending a session takes two steps that must not be separated:
// revoking the refresh-token rows stops future refreshes, but an access token
// already in flight is validated against the session cache, which never consults
// those rows. Bumping the per-user session version is what invalidates the
// cached entry.
//
// Leaving both steps to each caller failed in practice — change-password and
// delete-account did both, password-reset did only the first, so a stolen access
// token survived a reset for the remaining life of the JWT.
type SessionInvalidator interface {
	InvalidateAllForUser(ctx context.Context, userID int64, now time.Time) error
}

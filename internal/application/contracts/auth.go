// internal/application/contracts/auth_service.go
package contracts

import "context"

type contextKey int

const AuthContextKey contextKey = iota

// AuthContext is what the rest of the app needs to know
// about the authenticated caller.
type AuthContext struct {
	UserID    string
	SessionID string

	// You can extend this later without breaking callers:
	// Email     string
	// OrgID     *string
	// Roles     []string
	// Scopes    []string
	// IsMFAOK   bool
}

type AuthService interface {
	// Authenticate validates the provided access token, checks that the
	// underlying session is still valid (not expired/revoked), and returns
	// a normalized AuthContext for downstream use.
	//
	// Returns an error (e.g. ErrUnauthenticated) if:
	//   - token is invalid/expired
	//   - session does not exist
	//   - session is revoked/expired
	Authenticate(ctx context.Context, accessToken string) (AuthContext, error)
	CheckUserAccess(ctx context.Context, userID int64) error
}

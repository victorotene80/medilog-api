package services

import (
	"context"
	"fmt"
	"time"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

var _ appContracts.SessionInvalidator = (*sessionInvalidator)(nil)

type sessionInvalidator struct {
	refreshRepo repository.RefreshTokenRepository
	version     appContracts.SessionCache
}

func NewSessionInvalidator(
	refreshRepo repository.RefreshTokenRepository,
	version appContracts.SessionCache,
) appContracts.SessionInvalidator {
	return &sessionInvalidator{refreshRepo: refreshRepo, version: version}
}

func (s *sessionInvalidator) InvalidateAllForUser(
	ctx context.Context,
	userID int64,
	now time.Time,
) error {
	if err := s.refreshRepo.RevokeAllForUser(ctx, userID, now); err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}

	if s.version == nil {
		return nil
	}

	// Not best-effort: without this bump the caller's already-issued access
	// tokens keep authenticating off the session cache.
	if err := s.version.IncrementVersion(ctx, userID); err != nil {
		return fmt.Errorf("increment session version: %w", err)
	}

	return nil
}

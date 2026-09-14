package dto

import (
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
)

// RefreshResultDTO carries an expiry on each token rather than one shared
// ExpiresAt. The single field held the *access* token's expiry while the HTTP
// layer reported it as the refresh token's, so clients re-authenticated daily
// instead of weekly.
type RefreshResultDTO struct {
	AccessToken  contracts.Token
	RefreshToken contracts.Token
}
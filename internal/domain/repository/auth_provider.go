package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type UserAuthProviderRepository interface {
	FindByProviderUID(ctx context.Context, provider, providerUID string) (*entities.UserAuthProvider, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.UserAuthProvider, error)
	Create(ctx context.Context, p *entities.UserAuthProvider) error
}

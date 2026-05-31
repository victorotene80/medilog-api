package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.User, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.User, error)

	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByPhone(ctx context.Context, phone string) (*entities.User, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)

	Create(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error
	DeleteByPublicID(ctx context.Context, publicID string) error
}

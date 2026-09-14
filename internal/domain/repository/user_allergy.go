package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type UserAllergyRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.UserAllergy, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.UserAllergy, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.UserAllergy, error)
	Save(ctx context.Context, allergy *entities.UserAllergy) error
	Update(ctx context.Context, allergy *entities.UserAllergy) error
	Delete(ctx context.Context, id int64) error
	DeleteByPublicID(ctx context.Context, userID int64, publicID string) error
	SaveBatch(ctx context.Context, allergies []*entities.UserAllergy) error
	FindByUserIDAndAllergyID(ctx context.Context, userID int64, allergyID int64) (*entities.UserAllergy, error)
	FindByUserIDAndName(ctx context.Context, userID int64, name string) (*entities.UserAllergy, error)
}

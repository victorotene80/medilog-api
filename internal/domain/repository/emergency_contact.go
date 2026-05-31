package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type EmergencyContactRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.EmergencyContact, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.EmergencyContact, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.EmergencyContact, error)
	FindByUserIDAndPhone(ctx context.Context, userID int64, phone string) (*entities.EmergencyContact, error)
	Save(ctx context.Context, contact *entities.EmergencyContact) error
	Update(ctx context.Context, contact *entities.EmergencyContact) error
	Delete(ctx context.Context, id int64) error
	DeleteByPublicID(ctx context.Context, userID int64, publicID string) error
}

package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type EmergencyContactRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.EmergencyContact, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.EmergencyContact, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.EmergencyContact, error)
	FindByUserIDAndPhone(ctx context.Context, userID int64, phone string) (*entities.EmergencyContact, error)
	Save(ctx context.Context, contact *entities.EmergencyContact) error
	Update(ctx context.Context, contact *entities.EmergencyContact) error
	// SetPrimary makes contactID the user's only primary contact, demoting any
	// other in the same transaction. It must stay two statements rather than a
	// single `SET is_primary = (id = ?)`: Postgres evaluates a non-deferrable
	// unique index per row, so a combined update can transiently collide with
	// the one-primary-per-user index depending on row order.
	SetPrimary(ctx context.Context, userID, contactID int64, now time.Time) error
	Delete(ctx context.Context, id int64) error
	DeleteByPublicID(ctx context.Context, userID int64, publicID string) error
}

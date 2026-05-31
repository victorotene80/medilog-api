package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type MedicationAdherenceLogRepository interface {
	FindByID(ctx context.Context, id int64) (*entities.MedicationAdherenceLog, error)
	FindByMedicationID(ctx context.Context, medicationID int64) ([]*entities.MedicationAdherenceLog, error)
	FindByUserIDAndDateRange(ctx context.Context, userID int64, from, to time.Time) ([]*entities.MedicationAdherenceLog, error)
	Save(ctx context.Context, log *entities.MedicationAdherenceLog) error
	Update(ctx context.Context, log *entities.MedicationAdherenceLog) error
}

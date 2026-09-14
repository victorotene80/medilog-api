package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type MedicationTimeRepository interface {
	FindByMedicationID(ctx context.Context, medicationID int64) ([]*entities.MedicationTime, error)
	SaveAll(ctx context.Context, times []*entities.MedicationTime) error
	DeleteByMedicationID(ctx context.Context, medicationID int64) error
	ReplaceTimes(ctx context.Context, medicationID int64, times []*entities.MedicationTime) error
}

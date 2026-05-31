package persistence

import (
	"context"
	"encoding/json"
	"time"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

var _ appContracts.AuditLogger = (*AuditLogRepository)(nil)

type AuditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Log(ctx context.Context, rec dto.AuditRecord) error {
	occurredAt := rec.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	var metaJSON json.RawMessage
	if rec.Metadata != nil {
		b, err := json.Marshal(rec.Metadata)
		if err != nil {
			return err
		}
		metaJSON = b
	}

	m := &models.AuditLogModel{
		Action:      string(rec.Action),
		UserID:      rec.UserID,
		ActorID:     rec.ActorID,
		SessionID:   rec.SessionID,
		IPAddress:   rec.IPAddress,
		UserAgent:   rec.UserAgent,
		CountryCode: rec.CountryCode,
		TargetID:    rec.TargetID,
		Metadata:    metaJSON,
		Success:     rec.Success,
		OccurredAt:  occurredAt,
	}

	return r.db.WithContext(ctx).Create(m).Error
}

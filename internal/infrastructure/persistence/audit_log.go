package persistence

import (
	"context"
	"encoding/json"
	"time"

	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
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

	return conn(ctx, r.db).Create(m).Error
}

func (r *AuditLogRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entities.AuditLog, error) {
	var logs []*models.AuditLogModel
	query := conn(ctx, r.db).Where("user_id = ?", userID).Order("occurred_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return auditLogsToEntities(logs), nil
}

func (r *AuditLogRepository) FindAll(ctx context.Context, limit, offset int) ([]*entities.AuditLog, error) {
	var logs []*models.AuditLogModel
	query := conn(ctx, r.db).Order("occurred_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return auditLogsToEntities(logs), nil
}

func (r *AuditLogRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := conn(ctx, r.db).Model(&models.AuditLogModel{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AuditLogRepository) CountAll(ctx context.Context) (int64, error) {
	var count int64
	if err := conn(ctx, r.db).Model(&models.AuditLogModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// auditLogsToEntities keeps the GORM model inside infrastructure. Metadata is
// decoded here rather than shipped as raw JSON so the wire shape does not depend
// on how the column happens to be stored.
func auditLogsToEntities(models []*models.AuditLogModel) []*entities.AuditLog {
	out := make([]*entities.AuditLog, 0, len(models))

	for _, m := range models {
		if m == nil {
			continue
		}

		var metadata map[string]any
		if len(m.Metadata) > 0 {
			if err := json.Unmarshal(m.Metadata, &metadata); err != nil {
				metadata = nil
			}
		}

		out = append(out, &entities.AuditLog{
			ID:          m.ID,
			Action:      m.Action,
			UserID:      m.UserID,
			ActorID:     m.ActorID,
			SessionID:   m.SessionID,
			IPAddress:   m.IPAddress,
			UserAgent:   m.UserAgent,
			CountryCode: m.CountryCode,
			TargetID:    m.TargetID,
			Metadata:    metadata,
			Success:     m.Success,
			OccurredAt:  m.OccurredAt,
			CreatedAt:   m.CreatedAt,
		})
	}

	return out
}

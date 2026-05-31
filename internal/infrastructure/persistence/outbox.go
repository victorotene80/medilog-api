package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"github.com/victorotene80/medilog-api/internal/infrastructure/persistence/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	OutboxStatusPending    = 1
	OutboxStatusInProgress = 2
	OutboxStatusSent       = 3
	OutboxStatusFailed     = 4
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) (*OutboxRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}

	return &OutboxRepository{
		db: db,
	}, nil
}

func (r *OutboxRepository) Add(ctx context.Context, envelope appmsg.Envelope) error {
	model, err := envelopeToOutboxEventModel(envelope)
	if err != nil {
		return fmt.Errorf("map envelope to outbox model: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	return nil
}

func (r *OutboxRepository) FetchUnprocessed(ctx context.Context, limit int) ([]appmsg.Envelope, error) {
	/*if limit <= 0 {
		limit = 50
	}

	var rows []models.OutboxEventModel

	err := r.db.WithContext(ctx).
		Where("status IN ?", []int{OutboxStatusPending, OutboxStatusFailed}).
		Order("occurred_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("fetch unprocessed outbox events: %w", err)
	}

	envelopes := make([]appmsg.Envelope, 0, len(rows))

	for _, row := range rows {
		env, err := outboxEventModelToEnvelope(row)
		if err != nil {
			return nil, fmt.Errorf("map outbox model to envelope: %w", err)
		}

		envelopes = append(envelopes, env)
	}*/

	return nil, nil
}

func (r *OutboxRepository) MarkInProgress(ctx context.Context, id string) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&models.OutboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":         OutboxStatusInProgress,
			"in_progress_at": now,
			"updated_at":     now,
		})

	if result.Error != nil {
		return fmt.Errorf("mark outbox event in progress: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *OutboxRepository) MarkSent(ctx context.Context, id string) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&models.OutboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":         OutboxStatusSent,
			"sent_at":        now,
			"in_progress_at": nil,
			"last_error":     nil,
			"updated_at":     now,
		})

	if result.Error != nil {
		return fmt.Errorf("mark outbox event sent: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id string) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&models.OutboxEventModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":         OutboxStatusFailed,
			"in_progress_at": nil,
			"attempts":       gorm.Expr("attempts + 1"),
			"updated_at":     now,
		})

	if result.Error != nil {
		return fmt.Errorf("mark outbox event failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *OutboxRepository) ReclaimStaleInProgress(
	ctx context.Context,
	olderThan time.Time,
	limit int,
) (int, error) {
	if limit <= 0 {
		limit = 50
	}

	now := time.Now().UTC()

	subQuery := r.db.
		Model(&models.OutboxEventModel{}).
		Select("id").
		Where("status = ?", OutboxStatusInProgress).
		Where("in_progress_at < ?", olderThan).
		Order("in_progress_at ASC").
		Limit(limit)

	result := r.db.WithContext(ctx).
		Model(&models.OutboxEventModel{}).
		Where("id IN (?)", subQuery).
		Updates(map[string]any{
			"status":         OutboxStatusPending,
			"in_progress_at": nil,
			"updated_at":     now,
		})

	if result.Error != nil {
		return 0, fmt.Errorf("reclaim stale outbox events: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

func envelopeToOutboxEventModel(envelope appmsg.Envelope) (*models.OutboxEventModel, error) {
	metadataBytes, err := json.Marshal(envelope.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	if envelope.Metadata == nil {
		metadataBytes = []byte(`{}`)
	}

	return &models.OutboxEventModel{
		ID:            envelope.ID,
		Name:          envelope.Name,
		Kind:          string(envelope.Kind),
		AggregateID:   envelope.AggregateID,
		AggregateType: envelope.AggregateType,
		OccurredAt:    envelope.OccurredAt,
		Payload:       datatypes.JSON(envelope.Payload),
		Metadata:      datatypes.JSON(metadataBytes),
		CorrelationID: envelope.CorrelationID,
		CausationID:   envelope.CausationID,
		Version:       envelope.Version,
		Status:        OutboxStatusPending,
		Attempts:      0,
	}, nil
}

func outboxEventModelToEnvelope(model models.OutboxEventModel) (appmsg.Envelope, error) {
	var metadata map[string]string

	if len(model.Metadata) > 0 {
		if err := json.Unmarshal(model.Metadata, &metadata); err != nil {
			return appmsg.Envelope{}, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	if metadata == nil {
		metadata = map[string]string{}
	}

	return appmsg.Envelope{
		ID:            model.ID,
		Name:          model.Name,
		Kind:          appmsg.Kind(model.Kind),
		AggregateID:   model.AggregateID,
		AggregateType: model.AggregateType,
		OccurredAt:    model.OccurredAt,
		Payload:       []byte(model.Payload),
		Metadata:      metadata,
		CorrelationID: model.CorrelationID,
		CausationID:   model.CausationID,
		Version:       model.Version,
	}, nil
}

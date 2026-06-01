package models

import (
	"encoding/json"
	"fmt"
	"time"

	appmsg "github.com/victorotene80/medilog-api/internal/application/messaging"
	"gorm.io/datatypes"
)

type OutboxEventModel struct {
	ID            string         `gorm:"column:id;primaryKey;size:50"`
	Name          string         `gorm:"column:name;size:150;not null"`
	Kind          string         `gorm:"column:kind;size:50;not null"`
	AggregateID   int64          `gorm:"column:aggregate_id;not null;index"`
	AggregateType string         `gorm:"column:aggregate_type;size:100;not null"`
	OccurredAt    time.Time      `gorm:"column:occurred_at;not null;index"`
	Payload       datatypes.JSON `gorm:"column:payload;type:jsonb;not null"`
	Metadata      datatypes.JSON `gorm:"column:metadata;type:jsonb;not null"`
	CorrelationID *string        `gorm:"column:correlation_id;size:100"`
	CausationID   *string        `gorm:"column:causation_id;size:100"`
	Version       int            `gorm:"column:version;not null"`

	Status       int        `gorm:"column:status;not null;index"`
	Attempts     int        `gorm:"column:attempts;not null"`
	LastError    *string    `gorm:"column:last_error"`
	InProgressAt *time.Time `gorm:"column:in_progress_at"`
	SentAt       *time.Time `gorm:"column:sent_at"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (OutboxEventModel) TableName() string {
	return "outbox_events"
}

func envelopeToOutboxEventModel(envelope appmsg.Envelope) (*OutboxEventModel, error) {
	metadataBytes, err := json.Marshal(envelope.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	if envelope.Metadata == nil {
		metadataBytes = []byte(`{}`)
	}

	return &OutboxEventModel{
		ID:            envelope.ID,
		Name:          envelope.Name,
		Kind:          string(envelope.Kind),
		AggregateID:   envelope.AggregateID,
		AggregateType: envelope.AggregateType,
		OccurredAt:    envelope.OccurredAt,
		Payload:       datatypes.JSON(envelope.Payload),
		Metadata:      datatypes.JSON(metadataBytes),
		CorrelationID: nullableString(envelope.CorrelationID),
		CausationID:   nullableString(envelope.CausationID),
		Version:       envelope.Version,
		Status:        1,
		Attempts:      0,
	}, nil
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

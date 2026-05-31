package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type SupportAttachmentRepository interface {
	FindByMessageID(ctx context.Context, messageID int64) ([]*entities.SupportAttachment, error)
	Save(ctx context.Context, attachment *entities.SupportAttachment) error
}

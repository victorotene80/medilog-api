package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type SupportMessageRepository interface {
	FindByTicketID(ctx context.Context, ticketID int64) ([]*entities.SupportMessage, error)
	Save(ctx context.Context, msg *entities.SupportMessage) error
}

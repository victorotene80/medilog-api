package repository

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

// AuditLogQueryRepository is the read side of the audit trail. The write side
// is application/contracts.AuditLogger, which the same adapter implements.
type AuditLogQueryRepository interface {
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entities.AuditLog, error)
	FindAll(ctx context.Context, limit, offset int) ([]*entities.AuditLog, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	CountAll(ctx context.Context) (int64, error)
}

package contracts

import (
	"context"
	"github.com/victorotene80/medilog-api/internal/application/dto"
)

type AuditLogger interface {
	Log(ctx context.Context, rec dto.AuditRecord) error
}
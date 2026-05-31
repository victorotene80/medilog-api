package contracts
/*
import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/messaging"
)

type OutboxRepository interface {
	Add(ctx context.Context, envelope messaging.Envelope) error
	FetchUnprocessed(ctx context.Context, limit int) ([]messaging.Envelope, error)
	MarkInProgress(ctx context.Context, id string) error
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
	ReclaimStaleInProgress(ctx context.Context, olderThan time.Time, limit int) (int, error)
}*/
package repository

import (
	"context"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
)

type OTPCodeRepository interface {
	FindLatestByRecipientAndPurpose(ctx context.Context, recipient, purpose string) (*entities.OTPCode, error)
	FindByID(ctx context.Context, id int64) (*entities.OTPCode, error)
	Save(ctx context.Context, otp *entities.OTPCode) error
	Update(ctx context.Context, otp *entities.OTPCode) error
	InvalidatePreviousByRecipientAndPurpose(ctx context.Context, recipient, purpose string, now time.Time) error
}

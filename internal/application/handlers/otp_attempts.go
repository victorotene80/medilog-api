package handlers

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	"go.uber.org/zap"
)

// recordFailedOTPAttempt counts a wrong guess against the code itself.
//
// The counter must live on the row. The route limiter keys on the client IP, so
// it bounds guesses per address rather than per code — and against a 10^6
// keyspace a pool of addresses works through a meaningful fraction of it well
// inside the code's TTL. Once entities.MaxOTPAttempts is reached IsValid returns
// false and the attacker has to request a fresh code, which is separately rate
// limited and invalidates the previous one.
//
// Persisting is best-effort on purpose: the caller is already returning an
// authentication failure, and propagating a write error here would hand an
// attacker a way to keep the counter at zero by making the update fail.
func recordFailedOTPAttempt(
	ctx context.Context,
	repo repository.OTPCodeRepository,
	otp *entities.OTPCode,
) {
	if otp == nil {
		return
	}

	otp.RecordFailedAttempt()

	if err := repo.Update(ctx, otp); err != nil {
		zap.L().Warn("could not record failed OTP attempt",
			zap.Int64("otp_id", otp.ID),
			zap.Error(err),
		)
	}
}

package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"
)

// GetAIQuotaHandler reports the caller's remaining AI questions.
//
// It is a read path that can still mutate: the daily window is reset lazily
// here as well as on the send path, which is what lets a daily quota work
// without a scheduler.
type GetAIQuotaHandler struct {
	profiles repository.UserProfileRepository
	clock    func() time.Time
}

func NewGetAIQuotaHandler(
	profiles repository.UserProfileRepository,
	clock func() time.Time,
) *GetAIQuotaHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &GetAIQuotaHandler{profiles: profiles, clock: clock}
}

func (h *GetAIQuotaHandler) Handle(
	ctx context.Context,
	q query.GetAIQuotaQuery,
) (*dto.AIQuotaDTO, error) {
	if q.UserID <= 0 {
		return nil, application.NewValidation("user id is required")
	}

	profile, err := h.profiles.FindByUserID(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user profile: %w", err)
	}

	if profile == nil {
		return nil, application.NewNotFound("user profile not found")
	}

	now := h.clock()

	if profile.EnsureDailyWindow(now) {
		if err := h.profiles.Update(ctx, profile); err != nil {
			return nil, fmt.Errorf("reset ai quota window: %w", err)
		}
	}

	result := AIQuotaToDTO(profile, now)

	return &result, nil
}

// AIQuotaToDTO is shared with the send-message response so the client sees the
// same shape whether it polls the quota or reads it off a reply.
func AIQuotaToDTO(profile *entities.UserProfile, now time.Time) dto.AIQuotaDTO {
	return dto.AIQuotaDTO{
		Used:      profile.AIQuestionsUsed,
		Total:     profile.AIQuestionsTotal,
		Remaining: profile.AIQuestionsRemaining(),
		IsPro:     profile.AIIsPro,
		ResetsAt:  profile.AIQuotaResetsAt(now),
	}
}

package handlers

import (
	clockpkg "github.com/victorotene80/medilog-api/internal/shared/clock"

	"context"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type SubmitFeedbackHandler struct {
	feedbackRepo repository.FeedbackRepository
	clock        func() time.Time
}

func NewSubmitFeedbackHandler(
	feedbackRepo repository.FeedbackRepository,
	clock func() time.Time,
) *SubmitFeedbackHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &SubmitFeedbackHandler{
		feedbackRepo: feedbackRepo,
		clock:        clock,
	}
}

func (h *SubmitFeedbackHandler) Handle(
	ctx context.Context,
	cmd command.SubmitFeedbackCommand,
) (*dto.FeedbackResponseDTO, error) {
	now := h.clock()

	feedback := &entities.Feedback{
		UserID:      cmd.UserID,
		Rating:      cmd.Rating,
		Title:       cmd.Title,
		Message:     cmd.Message,
		AppVersion:  cmd.AppVersion,
		Platform:    cmd.Platform,
		DeviceModel: cmd.DeviceModel,
		Status:      "pending",
		CreatedAt:   now,
	}

	if err := h.feedbackRepo.Save(ctx, feedback); err != nil {
		return nil, fmt.Errorf("save feedback: %w", err)
	}

	return feedbackToDTO(feedback), nil
}

func feedbackToDTO(f *entities.Feedback) *dto.FeedbackResponseDTO {
	if f == nil {
		return nil
	}
	return &dto.FeedbackResponseDTO{
		ID:          f.PublicID,
		Rating:      f.Rating,
		Title:       f.Title,
		Message:     f.Message,
		AppVersion:  f.AppVersion,
		Platform:    f.Platform,
		DeviceModel: f.DeviceModel,
		Status:      f.Status,
		CreatedAt:   f.CreatedAt,
	}
}

type GetFeedbackHandler struct {
	feedbackRepo repository.FeedbackRepository
	clock        func() time.Time
}

func NewGetFeedbackHandler(
	feedbackRepo repository.FeedbackRepository,
	clock func() time.Time,
) *GetFeedbackHandler {
	if clock == nil {
		clock = clockpkg.Default()
	}
	return &GetFeedbackHandler{
		feedbackRepo: feedbackRepo,
		clock:        clock,
	}
}

func (h *GetFeedbackHandler) Handle(
	ctx context.Context,
	cmd command.GetFeedbackQuery,
) (*dto.FeedbackResponseDTO, error) {
	feedback, err := h.feedbackRepo.FindByID(ctx, cmd.FeedbackID)
	if err != nil {
		return nil, fmt.Errorf("find feedback: %w", err)
	}

	if feedback == nil {
		return nil, application.NewNotFound("feedback not found")
	}

	// Fail closed. Anonymous feedback has a nil UserID, and the previous
	// conjunction short-circuited on it — so every anonymous submission was
	// readable by any signed-in caller walking the sequential ids. Not-found
	// rather than forbidden, so the response cannot confirm the id exists.
	if cmd.UserID == nil || feedback.UserID == nil || *feedback.UserID != *cmd.UserID {
		return nil, application.NewNotFound("feedback not found")
	}

	return feedbackToDTO(feedback), nil
}

package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/mapper"
	"github.com/victorotene80/medilog-api/internal/application/query"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type GetDashboardHandler struct {
	dashboard domainRepo.DashboardRepository
	clock     func() time.Time
}

func NewGetDashboardHandler(
	dashboard domainRepo.DashboardRepository,
	clock func() time.Time,
) *GetDashboardHandler {
	return &GetDashboardHandler{dashboard: dashboard, clock: clock}
}

func (h *GetDashboardHandler) Handle(
	ctx context.Context,
	q query.GetDashboardQuery,
) (*dto.DashboardDTO, error) {
	if q.UserID <= 0 {
		return nil, errors.New("user id is required")
	}

	dashboard, err := h.dashboard.GetByUserID(ctx, q.UserID, h.clock())
	if err != nil {
		return nil, fmt.Errorf("get dashboard: %w", err)
	}

	return mapper.DashboardToDTO(dashboard), nil
}

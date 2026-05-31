package handlers

import (
	"context"
	"fmt"

	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	domainRepo "github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListDrugScansHandler struct {
	drugScans domainRepo.DrugScanRepository
}

func NewListDrugScansHandler(drugScans domainRepo.DrugScanRepository) *ListDrugScansHandler {
	return &ListDrugScansHandler{drugScans: drugScans}
}

func (h *ListDrugScansHandler) Handle(
	ctx context.Context,
	q query.ListDrugScansQuery,
) ([]dto.DrugScanDTO, error) {
	scans, err := h.drugScans.FindByUserID(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("list drug scans: %w", err)
	}

	result := make([]dto.DrugScanDTO, 0, len(scans))
	for _, s := range scans {
		result = append(result, toScanDTO(s, nil)) // no medicine hydration on list
	}
	return result, nil
}

type GetDrugScanHandler struct {
	drugScans domainRepo.DrugScanRepository
}

func NewGetDrugScanHandler(drugScans domainRepo.DrugScanRepository) *GetDrugScanHandler {
	return &GetDrugScanHandler{drugScans: drugScans}
}

func (h *GetDrugScanHandler) Handle(
	ctx context.Context,
	q query.GetDrugScanQuery,
) (*dto.DrugScanDTO, error) {
	scan, err := h.drugScans.FindByPublicID(ctx, q.UserID, q.PublicID)
	if err != nil {
		return nil, fmt.Errorf("get drug scan: %w", err)
	}
	if scan == nil {
		return nil, nil
	}
	d := toScanDTO(scan, nil)
	return &d, nil
}

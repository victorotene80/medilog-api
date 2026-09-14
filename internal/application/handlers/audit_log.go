package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

type ListAuditLogsHandler struct {
	auditLogs repository.AuditLogQueryRepository
}

func NewListAuditLogsHandler(
	auditLogs repository.AuditLogQueryRepository,
) *ListAuditLogsHandler {
	return &ListAuditLogsHandler{auditLogs: auditLogs}
}

// Handle lists audit entries, optionally scoped to one user.
//
// The route is admin-gated in the router, so there is no per-caller branch here:
// the handler previously took an IsAdmin flag that the REST layer derived from
// "is authenticated", which is not the same question and made the flag a lie
// about what it measured.
func (h *ListAuditLogsHandler) Handle(
	ctx context.Context,
	cmd command.ListAuditLogsQuery,
) (*dto.ListAuditLogsDTO, error) {
	limit := cmd.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := cmd.Offset
	if offset < 0 {
		offset = 0
	}

	var (
		logs  []*entities.AuditLog
		total int64
		err   error
	)

	if cmd.UserID == "" {
		if logs, err = h.auditLogs.FindAll(ctx, limit, offset); err != nil {
			return nil, fmt.Errorf("find audit logs: %w", err)
		}
		if total, err = h.auditLogs.CountAll(ctx); err != nil {
			return nil, fmt.Errorf("count audit logs: %w", err)
		}
	} else {
		if logs, err = h.auditLogs.FindByUserID(ctx, cmd.UserID, limit, offset); err != nil {
			return nil, fmt.Errorf("find audit logs: %w", err)
		}
		if total, err = h.auditLogs.CountByUserID(ctx, cmd.UserID); err != nil {
			return nil, fmt.Errorf("count audit logs: %w", err)
		}
	}

	items := make([]*dto.AuditLogDTO, 0, len(logs))
	for _, l := range logs {
		items = append(items, auditLogToDTO(l))
	}

	return &dto.ListAuditLogsDTO{
		Logs:   items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func auditLogToDTO(l *entities.AuditLog) *dto.AuditLogDTO {
	if l == nil {
		return nil
	}

	return &dto.AuditLogDTO{
		ID:          strconv.FormatInt(l.ID, 10),
		Action:      l.Action,
		UserID:      l.UserID,
		ActorID:     l.ActorID,
		SessionID:   l.SessionID,
		IPAddress:   l.IPAddress,
		UserAgent:   l.UserAgent,
		CountryCode: l.CountryCode,
		TargetID:    l.TargetID,
		Metadata:    l.Metadata,
		Success:     l.Success,
		OccurredAt:  l.OccurredAt,
		CreatedAt:   l.CreatedAt,
	}
}

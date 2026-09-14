package handlers

import (
	"context"

	"github.com/victorotene80/medilog-api/internal/application/command"
	appContracts "github.com/victorotene80/medilog-api/internal/application/contracts"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/domain/contracts"
	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

type RefreshSessionHandler struct {
	sessionService appContracts.SessionService
}

func NewRefreshSessionHandler(
	sessionService appContracts.SessionService,
) *RefreshSessionHandler {
	return &RefreshSessionHandler{
		sessionService: sessionService,
	}
}

func (h *RefreshSessionHandler) Handle(
	ctx context.Context,
	cmd command.RefreshSessionCommand,
) (*dto.RefreshResultDTO, error) {
	meta, _ := requestmeta.FromContext(ctx)

	sessionResult, err := h.sessionService.Refresh(
		ctx,
		cmd.RefreshToken,
		meta.IPAddress,
		meta.UserAgent,
		cmd.DeviceID,
		cmd.DeviceFingerprint,
		cmd.DeviceName,
	)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshResultDTO{
		AccessToken: contracts.Token{
			Value:     sessionResult.AccessToken.Value,
			ExpiresAt: sessionResult.AccessToken.ExpiresAt,
		},
		RefreshToken: contracts.Token{
			Value:     sessionResult.RefreshToken.Value,
			ExpiresAt: sessionResult.RefreshToken.ExpiresAt,
		},
	}, nil
}

package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/victorotene80/medilog-api/internal/application"
	"github.com/victorotene80/medilog-api/internal/application/command"
	"github.com/victorotene80/medilog-api/internal/application/dto"
	"github.com/victorotene80/medilog-api/internal/application/query"
	"github.com/victorotene80/medilog-api/internal/domain/entities"
	"github.com/victorotene80/medilog-api/internal/domain/repository"
)

// GetNotificationPreferencesHandler reads the caller's delivery preferences.
type GetNotificationPreferencesHandler struct {
	profiles repository.UserProfileRepository
}

func NewGetNotificationPreferencesHandler(
	profiles repository.UserProfileRepository,
) *GetNotificationPreferencesHandler {
	return &GetNotificationPreferencesHandler{profiles: profiles}
}

func (h *GetNotificationPreferencesHandler) Handle(
	ctx context.Context,
	q query.GetNotificationPreferencesQuery,
) (*dto.NotificationPreferencesDTO, error) {
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

	result := NotificationPreferencesToDTO(profile)

	return &result, nil
}

// UpdateNotificationPreferencesHandler applies a partial update to the caller's
// delivery preferences. Every field is optional; nil means "leave unchanged".
type UpdateNotificationPreferencesHandler struct {
	profiles repository.UserProfileRepository
}

func NewUpdateNotificationPreferencesHandler(
	profiles repository.UserProfileRepository,
) *UpdateNotificationPreferencesHandler {
	return &UpdateNotificationPreferencesHandler{profiles: profiles}
}

func (h *UpdateNotificationPreferencesHandler) Handle(
	ctx context.Context,
	cmd command.UpdateNotificationPreferencesCommand,
) (*dto.NotificationPreferencesDTO, error) {
	if cmd.UserID <= 0 {
		return nil, application.NewValidation("user id is required")
	}

	profile, err := h.profiles.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user profile: %w", err)
	}

	if profile == nil {
		return nil, application.NewNotFound("user profile not found")
	}

	// Validate before mutating so a bad timezone cannot leave half the toggles
	// applied.
	if cmd.Timezone != nil {
		tz := strings.TrimSpace(*cmd.Timezone)
		if tz == "" {
			return nil, application.NewValidation("timezone cannot be empty")
		}

		// The scheduler calls time.LoadLocation on this value on every tick.
		// Rejecting it here is what keeps an unresolvable zone out of the table.
		if _, err := time.LoadLocation(tz); err != nil {
			return nil, application.NewValidation(
				"timezone must be a valid IANA name, for example Africa/Lagos",
			)
		}

		profile.Timezone = tz
	}

	applyBool(&profile.MedicationRemindersEnabled, cmd.MedicationRemindersEnabled)
	applyBool(&profile.RefillRemindersEnabled, cmd.RefillRemindersEnabled)
	applyBool(&profile.AppointmentRemindersEnabled, cmd.AppointmentRemindersEnabled)
	applyBool(&profile.AIHealthTipsEnabled, cmd.AIHealthTipsEnabled)
	applyBool(&profile.SupportUpdatesEnabled, cmd.SupportUpdatesEnabled)
	applyBool(&profile.AppUpdatesEnabled, cmd.AppUpdatesEnabled)
	applyBool(&profile.PushEnabled, cmd.PushEnabled)
	applyBool(&profile.EmailEnabled, cmd.EmailEnabled)
	applyBool(&profile.SMSEnabled, cmd.SMSEnabled)
	applyBool(&profile.WhatsAppEnabled, cmd.WhatsAppEnabled)

	if err := h.profiles.Update(ctx, profile); err != nil {
		return nil, fmt.Errorf("update notification preferences: %w", err)
	}

	result := NotificationPreferencesToDTO(profile)

	return &result, nil
}

func applyBool(target *bool, value *bool) {
	if value != nil {
		*target = *value
	}
}

func NotificationPreferencesToDTO(p *entities.UserProfile) dto.NotificationPreferencesDTO {
	return dto.NotificationPreferencesDTO{
		MedicationRemindersEnabled:  p.MedicationRemindersEnabled,
		RefillRemindersEnabled:      p.RefillRemindersEnabled,
		AppointmentRemindersEnabled: p.AppointmentRemindersEnabled,
		AIHealthTipsEnabled:         p.AIHealthTipsEnabled,
		SupportUpdatesEnabled:       p.SupportUpdatesEnabled,
		AppUpdatesEnabled:           p.AppUpdatesEnabled,
		PushEnabled:                 p.PushEnabled,
		EmailEnabled:                p.EmailEnabled,
		SMSEnabled:                  p.SMSEnabled,
		WhatsAppEnabled:             p.WhatsAppEnabled,
		Timezone:                    p.Timezone,
	}
}

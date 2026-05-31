package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/victorotene80/medilog-api/internal/domain/services/policy"
)

type MFAService struct {
	policy policy.MFAPolicy
}

func NewMFAService(p policy.MFAPolicy) *MFAService {
	return &MFAService{policy: p}
}

func (s *MFAService) IsMFARequiredForAction(action string) bool {
	if !s.policy.Required {
		return false
	}
	for _, a := range s.policy.RequiredActions {
		if a == action {
			return true
		}
	}
	return false
}

func (s *MFAService) IsWithinGracePeriod(lastValidated, now time.Time) bool {
	return now.Sub(lastValidated) <= s.policy.GracePeriod
}

func (s *MFAService) IsDeviceRemembered(rememberedAt, now time.Time) bool {
	if !s.policy.RememberDevice {
		return false
	}
	return now.Sub(rememberedAt) <= s.policy.RememberDuration
}

func (s *MFAService) StartMFAChallenge() (string, error) {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		return "", errors.New("failed to generate MFA challenge")
	}
	return hex.EncodeToString(id), nil
}

func (s *MFAService) ValidateMFAMethod(method policy.MFAMethod) bool {
	for _, m := range s.policy.AllowedMethods {
		if m == method {
			return true
		}
	}
	return false
}

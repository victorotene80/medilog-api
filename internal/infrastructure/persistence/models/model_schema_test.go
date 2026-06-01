package models

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestGormModelsParse(t *testing.T) {
	models := []any{
		&AIConversationModel{},
		&AIMessageModel{},
		&AllergyModel{},
		&AuditLogModel{},
		&CountryModel{},
		&DrugScanModel{},
		&EmergencyContactModel{},
		&FeedbackModel{},
		&FunFactModel{},
		&MedicationAdherenceLogModel{},
		&MedicationModel{},
		&MedicationTimeModel{},
		&NotificationModel{},
		&OTPCodeModel{},
		&OutboxEventModel{},
		&RefreshTokenModel{},
		&RegisteredMedicineModel{},
		&SupportAttachmentModel{},
		&SupportMessageModel{},
		&SupportTicketModel{},
		&UserAuthProviderModel{},
		&UserAllergyModel{},
		&UserModel{},
		&UserProfileModel{},
		&VisitModel{},
	}

	for _, model := range models {
		model := model
		t.Run(schemaName(t, model), func(t *testing.T) {
			if _, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{}); err != nil {
				t.Fatalf("parse GORM model: %v", err)
			}
		})
	}
}

func schemaName(t *testing.T, model any) string {
	t.Helper()

	parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return "unparseable"
	}
	return parsed.Name
}

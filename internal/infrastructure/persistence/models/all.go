package models

// All returns every GORM model the API persists. It is the single source of
// truth for schema creation (DB_AUTO_MIGRATE=true), startup verification
// (DB_VERIFY_SCHEMA=true) and model parsing tests.
func All() []any {
	return []any{
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
}

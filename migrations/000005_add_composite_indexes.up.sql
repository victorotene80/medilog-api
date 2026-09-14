-- Composite indexes for common query patterns

-- Medications: FindByUserID, FindActiveByUserID
CREATE INDEX IF NOT EXISTS idx_medications_user_id_deleted_at ON medications(user_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_medications_user_id_completed ON medications(user_id, is_completed, deleted_at);

-- Medication adherence logs: FindByMedicationID, FindByUserIDAndDateRange
CREATE INDEX IF NOT EXISTS idx_adherence_logs_medication_id_deleted ON medication_adherence_logs(medication_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_adherence_logs_user_scheduled ON medication_adherence_logs(user_id, scheduled_at, deleted_at);

-- Visits: FindByUserID, FindByUserIDAndDateRange
CREATE INDEX IF NOT EXISTS idx_visits_user_id_deleted_at ON visits(user_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_visits_user_id_date ON visits(user_id, visit_date, deleted_at);

-- Drug scans: FindByUserID
CREATE INDEX IF NOT EXISTS idx_drug_scans_user_id_deleted_at ON drug_scans(user_id, deleted_at);

-- AI conversations: FindActiveByUserID
CREATE INDEX IF NOT EXISTS idx_ai_conversations_user_id_deleted_at ON ai_conversations(user_id, deleted_at);

-- Notifications: FindByUserID, FindUnreadByUserID
CREATE INDEX IF NOT EXISTS idx_notifications_user_id_deleted_at ON notifications(user_id, deleted_at);

-- Support tickets: FindByUserID
CREATE INDEX IF NOT EXISTS idx_support_tickets_user_id_deleted_at ON support_tickets(user_id, deleted_at);

-- User allergies: FindByUserID
CREATE INDEX IF NOT EXISTS idx_user_allergies_user_id_deleted_at ON user_allergies(user_id, deleted_at);

-- Emergency contacts: FindByUserID
CREATE INDEX IF NOT EXISTS idx_emergency_contacts_user_id_deleted_at ON emergency_contacts(user_id, deleted_at);

-- OTP codes: FindLatestByRecipientAndPurpose
CREATE INDEX IF NOT EXISTS idx_otp_codes_recipient_purpose_created ON otp_codes(recipient, purpose, created_at);

-- Refresh token: FindByTokenHash, FindActiveByUserID
CREATE INDEX IF NOT EXISTS idx_refresh_token_token_hash ON refresh_token(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_token_user_id_revoked_expires ON refresh_token(user_id, revoked_at, expires_at);

-- Outbox events: FetchUnprocessed
CREATE INDEX IF NOT EXISTS idx_outbox_events_status_occurred ON outbox_events(status, occurred_at);

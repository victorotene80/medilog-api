-- Remove composite indexes

DROP INDEX IF EXISTS idx_medications_user_id_deleted_at;
DROP INDEX IF EXISTS idx_medications_user_id_completed;
DROP INDEX IF EXISTS idx_adherence_logs_medication_id_deleted;
DROP INDEX IF EXISTS idx_adherence_logs_user_scheduled;
DROP INDEX IF EXISTS idx_visits_user_id_deleted_at;
DROP INDEX IF EXISTS idx_visits_user_id_date;
DROP INDEX IF EXISTS idx_drug_scans_user_id_deleted_at;
DROP INDEX IF EXISTS idx_ai_conversations_user_id_deleted_at;
DROP INDEX IF EXISTS idx_notifications_user_id_deleted_at;
DROP INDEX IF EXISTS idx_support_tickets_user_id_deleted_at;
DROP INDEX IF EXISTS idx_user_allergies_user_id_deleted_at;
DROP INDEX IF EXISTS idx_emergency_contacts_user_id_deleted_at;
DROP INDEX IF EXISTS idx_otp_codes_recipient_purpose_created;
DROP INDEX IF EXISTS idx_refresh_token_token_hash;
DROP INDEX IF EXISTS idx_refresh_token_user_id_revoked_expires;
DROP INDEX IF EXISTS idx_outbox_events_status_occurred;

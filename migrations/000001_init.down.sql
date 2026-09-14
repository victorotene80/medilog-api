-- Rollback initial schema
-- Drop all tables in reverse order of creation

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS feedback;
DROP TABLE IF EXISTS support_attachments;
DROP TABLE IF EXISTS support_messages;
DROP TABLE IF EXISTS support_tickets;
DROP TABLE IF EXISTS ai_messages;
DROP TABLE IF EXISTS ai_conversations;
DROP TABLE IF EXISTS fun_facts;
DROP TABLE IF EXISTS countries;
DROP TABLE IF EXISTS registered_medicines;
DROP TABLE IF EXISTS drug_scans;
DROP TABLE IF EXISTS emergency_contacts;
DROP TABLE IF EXISTS user_allergies;
DROP TABLE IF EXISTS allergies;
DROP TABLE IF EXISTS visits;
DROP TABLE IF EXISTS medication_adherence_logs;
DROP TABLE IF EXISTS medication_times;
DROP TABLE IF EXISTS medications;
DROP TABLE IF EXISTS otp_codes;
DROP TABLE IF EXISTS refresh_token;
DROP TABLE IF EXISTS user_auth_providers;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;

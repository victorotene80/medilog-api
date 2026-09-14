-- Ensure emergency_contacts.country_code exists.
--
-- This migration originally ran an unconditional
--     ADD COLUMN country_code VARCHAR(4) NOT NULL DEFAULT ''
-- but 000001_init.up.sql already declares country_code TEXT NOT NULL on this
-- table, so it aborted with SQLSTATE 42701 (duplicate column) on every clean
-- database. IF NOT EXISTS makes it idempotent while keeping its stated intent.
--
-- Editing this file in place is safe and is the only thing that can work:
-- golang-migrate records only (version, dirty) in schema_migrations and does
-- not checksum migration files, so already-migrated databases never re-read
-- it. A forward "repair" migration could not help, because 000002 fails before
-- the runner ever reaches a later version.
--
-- The column type stays TEXT, matching 000001 and the untagged string field on
-- models.EmergencyContactModel.
ALTER TABLE emergency_contacts
    ADD COLUMN IF NOT EXISTS country_code TEXT NOT NULL DEFAULT '';

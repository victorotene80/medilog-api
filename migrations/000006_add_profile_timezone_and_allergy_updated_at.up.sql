-- Per-user IANA timezone for reminder scheduling, and a modification timestamp
-- for user_allergies now that PUT /health/allergies/{publicId} can edit a
-- record in place.
--
-- timezone lives on user_profiles rather than users because it is a preference
-- alongside the notification flags already stored here, and the reminder
-- scheduler joins this table anyway. It is NOT derived from users.country_code:
-- that works for single-zone countries and breaks for the US, Brazil, Russia,
-- Australia and Canada.
--
-- 'UTC' is a safe default: every existing row becomes immediately valid and the
-- scheduler always has a resolvable zone. Clients overwrite it with the device
-- zone via PATCH /users/me/notification-preferences.

ALTER TABLE user_profiles
    ADD COLUMN IF NOT EXISTS timezone TEXT NOT NULL DEFAULT 'UTC';

ALTER TABLE user_allergies
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

-- Existing rows have never been edited, so their modification time is their
-- creation time.
UPDATE user_allergies SET updated_at = created_at WHERE updated_at IS NULL;

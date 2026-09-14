-- Server-generated reminders need to be idempotent: the scheduler re-scans an
-- overlapping window on every tick, so it must be able to insert "the 08:00
-- dose on 2026-09-08" repeatedly without ever producing a second row.

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS dedupe_key TEXT;

-- Deliberately NOT a partial index.
--
-- Postgres treats NULLs as distinct in a unique index, so every hand-authored
-- notification (dedupe_key IS NULL) is unaffected and unlimited. Keeping the
-- target a plain two-column pair also lets ON CONFLICT (user_id, dedupe_key)
-- infer it directly; a partial index would need its predicate repeated in the
-- conflict target, which is easy to get subtly wrong.
--
-- It also deliberately does NOT exclude soft-deleted rows: a reminder the user
-- dismissed must not reappear on the next tick.
CREATE UNIQUE INDEX IF NOT EXISTS ux_notifications_user_dedupe
    ON notifications (user_id, dedupe_key);

-- Inbox listing: newest first, keyset-paginated on (created_at, id).
CREATE INDEX IF NOT EXISTS idx_notifications_user_created_id
    ON notifications (user_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

-- Unread badge count.
CREATE INDEX IF NOT EXISTS idx_notifications_user_status_unread
    ON notifications (user_id)
    WHERE deleted_at IS NULL AND status = 'unread';

-- Scheduler scan paths. The scheduler sweeps globally by date rather than by
-- user, so the existing user-scoped visit index does not serve it.
CREATE INDEX IF NOT EXISTS idx_medications_reminder_scan
    ON medications (id)
    WHERE deleted_at IS NULL AND is_completed = false;

CREATE INDEX IF NOT EXISTS idx_medication_times_med_deleted
    ON medication_times (medication_id, deleted_at);

CREATE INDEX IF NOT EXISTS idx_visits_visit_date_deleted
    ON visits (visit_date, deleted_at);

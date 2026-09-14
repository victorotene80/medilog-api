DROP INDEX IF EXISTS idx_visits_visit_date_deleted;
DROP INDEX IF EXISTS idx_medication_times_med_deleted;
DROP INDEX IF EXISTS idx_medications_reminder_scan;
DROP INDEX IF EXISTS idx_notifications_user_status_unread;
DROP INDEX IF EXISTS idx_notifications_user_created_id;
DROP INDEX IF EXISTS ux_notifications_user_dedupe;

ALTER TABLE notifications DROP COLUMN IF EXISTS dedupe_key;

DROP INDEX IF EXISTS idx_feedback_deleted_at;
DROP INDEX IF EXISTS idx_notifications_deleted_at;
DROP INDEX IF EXISTS idx_support_attachments_deleted_at;
DROP INDEX IF EXISTS idx_support_messages_deleted_at;
DROP INDEX IF EXISTS idx_support_tickets_deleted_at;

ALTER TABLE feedback DROP COLUMN deleted_at;
ALTER TABLE notifications DROP COLUMN deleted_at;
ALTER TABLE support_attachments DROP COLUMN deleted_at;
ALTER TABLE support_messages DROP COLUMN deleted_at;
ALTER TABLE support_tickets DROP COLUMN deleted_at;

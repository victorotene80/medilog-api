ALTER TABLE support_tickets ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE support_messages ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE support_attachments ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE notifications ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE feedback ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_support_tickets_deleted_at ON support_tickets(deleted_at);
CREATE INDEX idx_support_messages_deleted_at ON support_messages(deleted_at);
CREATE INDEX idx_support_attachments_deleted_at ON support_attachments(deleted_at);
CREATE INDEX idx_notifications_deleted_at ON notifications(deleted_at);
CREATE INDEX idx_feedback_deleted_at ON feedback(deleted_at);

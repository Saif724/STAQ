CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    execution_id UUID REFERENCES executions(id) ON DELETE CASCADE,
    action_id UUID,
    type VARCHAR(30) NOT NULL DEFAULT 'REMINDER',
    message TEXT NOT NULL,
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT uq_notifications_execution_action UNIQUE (execution_id, action_id)
);

CREATE INDEX idx_notifications_user_created
    ON notifications (user_id, created_at DESC);

CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id) WHERE read_at IS NULL;
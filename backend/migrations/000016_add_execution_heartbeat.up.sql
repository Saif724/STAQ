ALTER TABLE executions
ADD COLUMN heartbeat_at TIMESTAMP NULL;

UPDATE executions
SET heartbeat_at = CURRENT_TIMESTAMP
WHERE status = 'RUNNING';

CREATE INDEX idx_executions_running_heartbeat
    ON executions(heartbeat_at)
    WHERE status = 'RUNNING';
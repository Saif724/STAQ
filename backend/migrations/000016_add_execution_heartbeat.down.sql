DROP INDEX IF EXISTS idx_executions_running_heartbeat;

ALTER TABLE executions
DROP COLUMN IF EXISTS heartbeat_at;
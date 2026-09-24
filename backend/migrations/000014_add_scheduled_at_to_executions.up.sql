ALTER TABLE executions
ADD COLUMN scheduled_at TIMESTAMP NULL;

UPDATE executions
SET scheduled_at = started_at
WHERE scheduled_at IS NULL;

ALTER TABLE executions
ALTER COLUMN scheduled_at SET NOT NULL;

ALTER TABLE executions
ADD CONSTRAINT uq_executions_trigger_scheduled
UNIQUE (trigger_id, scheduled_at);
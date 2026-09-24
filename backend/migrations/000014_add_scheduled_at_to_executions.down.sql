ALTER TABLE executions
DROP CONSTRAINT uq_executions_trigger_scheduled;

ALTER TABLE executions
DROP COLUMN scheduled_at;
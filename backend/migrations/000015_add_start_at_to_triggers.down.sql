DROP INDEX IF EXISTS idx_triggers_start_at;

ALTER TABLE triggers
DROP COLUMN IF EXISTS start_at;
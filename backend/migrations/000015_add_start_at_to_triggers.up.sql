ALTER TABLE triggers
ADD COLUMN start_at TIMESTAMP;

UPDATE triggers
SET start_at = next_run_at
WHERE start_at IS NULL;

ALTER TABLE triggers
ALTER COLUMN start_at SET NOT NULL;

CREATE INDEX idx_triggers_start_at
    ON triggers(start_at);
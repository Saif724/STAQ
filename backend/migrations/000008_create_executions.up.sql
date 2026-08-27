CREATE TABLE executions (
    id UUID PRIMARY KEY,
    task_id UUID NOT NULL,
    trigger_id UUID NOT NULL,
    status VARCHAR(30) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NULL,
    duration_ms BIGINT NULL,
    retry_count INTEGER DEFAULT 0,
    error_message TEXT NULL,
    created_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_executions_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id),

    CONSTRAINT fk_executions_trigger
        FOREIGN KEY (trigger_id)
        REFERENCES triggers(id),

    CONSTRAINT chk_executions_status
        CHECK (
            status IN (
                'PENDING',
                'RUNNING',
                'SUCCESS',
                'FAILED',
                'CANCELLED',
                'TIMED_OUT'
            )
        )
);

CREATE INDEX idx_executions_task
    ON executions(task_id);

CREATE INDEX idx_executions_status
    ON executions(status);

CREATE INDEX idx_executions_started
    ON executions(started_at);

CREATE INDEX idx_executions_trigger
    ON executions(trigger_id);
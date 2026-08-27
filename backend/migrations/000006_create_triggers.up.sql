CREATE TABLE triggers (
    id UUID PRIMARY KEY,
    task_id UUID NOT NULL,
    trigger_type VARCHAR(30) NOT NULL,
    cron_expression VARCHAR(100) NULL,
    timezone VARCHAR(50) NOT NULL,
    next_run_at TIMESTAMP NOT NULL,
    last_run_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_triggers_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id),

    CONSTRAINT chk_triggers_type
        CHECK (
            trigger_type IN (
                'ONCE',
                'DAILY',
                'WEEKLY',
                'MONTHLY',
                'YEARLY',
                'CRON'
            )
        )
);

CREATE INDEX idx_triggers_task
    ON triggers(task_id);

CREATE INDEX idx_triggers_next_run
    ON triggers(next_run_at);

CREATE INDEX idx_triggers_active
    ON triggers(is_active);
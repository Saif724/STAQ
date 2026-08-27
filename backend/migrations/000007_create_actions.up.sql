CREATE TABLE actions (
    id UUID PRIMARY KEY,
    task_id UUID NOT NULL,
    action_type VARCHAR(30) NOT NULL,
    execution_order INTEGER NOT NULL,
    configuration JSONB NOT NULL,
    continue_on_failure BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_actions_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id),

    CONSTRAINT uq_actions_task_order
        UNIQUE (task_id, execution_order),

    CONSTRAINT chk_actions_type
        CHECK (
            action_type IN (
                'REMINDER',
                'EMAIL',
                'HTTP',
                'SHELL'
            )
        )
);

CREATE INDEX idx_actions_task
    ON actions(task_id);

CREATE INDEX idx_actions_order
    ON actions(execution_order);
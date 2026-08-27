CREATE TABLE execution_logs (
    id UUID PRIMARY KEY,
    execution_id UUID NOT NULL,
    log_level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_execution_logs_execution
        FOREIGN KEY (execution_id)
        REFERENCES executions(id),
    
    CONSTRAINT chk_execution_logs_level
        CHECK (
            log_level IN (
                'INFO',
                'WARNING',
                'ERROR'
            )
        )
);

CREATE INDEX idx_logs_execution
    ON execution_logs(execution_id);

CREATE INDEX idx_logs_level
    ON execution_logs(log_level);

CREATE INDEX idx_logs_created
    ON execution_logs(created_at);
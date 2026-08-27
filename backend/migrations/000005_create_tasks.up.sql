CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    queue_id UUID NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT NULL,
    status VARCHAR(20) NOT NULL,
    timeout_seconds INTEGER DEFAULT 300,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_tasks_user
        FOREIGN KEY (user_id)
        REFERENCES users(id),

    CONSTRAINT fk_tasks_queue
        FOREIGN KEY (queue_id)
        REFERENCES queues(id),
    
    CONSTRAINT chk_tasks_status
        CHECK (status IN ('Active', 'Paused', 'Archived'))
);

CREATE INDEX idx_tasks_user
    ON tasks(user_id);

CREATE INDEX idx_tasks_queue
    ON tasks(queue_id);

CREATE INDEX idx_tasks_status
    ON tasks(status);

CREATE INDEX idx_tasks_created
    ON tasks(created_at);
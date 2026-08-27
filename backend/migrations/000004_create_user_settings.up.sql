CREATE TABLE user_settings (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    default_queue_id UUID NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_user_settings_user
        FOREIGN KEY (user_id)
        REFERENCES users(id),

    CONSTRAINT fk_user_settings_default_queue
        FOREIGN KEY (default_queue_id)
        REFERENCES queues(id)
);

CREATE INDEX idx_user_settings_user
    ON user_settings(user_id);
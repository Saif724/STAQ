CREATE TABLE email_verifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_email_verifications_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_email_verifications_token
    ON email_verifications(token);

CREATE INDEX idx_email_verifications_user_id
    ON email_verifications(user_id);

CREATE INDEX idx_email_verifications_expires_at
    ON email_verifications(expires_at);
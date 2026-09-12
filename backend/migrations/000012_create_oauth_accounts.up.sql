CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    provider VARCHAR(255) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CONSTRAINT fk_oauth_accounts_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    
    CONSTRAINT uq_oauth_provider_user
        UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_oauth_accounts_user_id
    ON oauth_accounts(user_id);

CREATE INDEX idx_oauth_accounts_provider
    ON oauth_accounts(provider);
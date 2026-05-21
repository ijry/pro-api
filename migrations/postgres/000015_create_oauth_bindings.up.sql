CREATE TABLE IF NOT EXISTS oauth_bindings (
    id            BIGINT       NOT NULL PRIMARY KEY,
    user_id       BIGINT       NOT NULL,
    provider      VARCHAR(16)  NOT NULL,
    provider_uid  VARCHAR(128) NOT NULL,
    email         VARCHAR(128) NOT NULL DEFAULT '',
    profile       JSONB,
    created_at    TIMESTAMPTZ  NOT NULL,
    updated_at    TIMESTAMPTZ  NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_oauth_provider_uid ON oauth_bindings (provider, provider_uid);
CREATE INDEX IF NOT EXISTS idx_oauth_user_provider ON oauth_bindings (user_id, provider);

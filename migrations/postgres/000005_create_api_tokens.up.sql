CREATE TABLE IF NOT EXISTS api_tokens (
    id             BIGINT       NOT NULL PRIMARY KEY,
    user_id        BIGINT       NOT NULL,
    name           VARCHAR(64)  NOT NULL DEFAULT '',
    key_hash       VARCHAR(64)  NOT NULL,
    key_prefix     VARCHAR(32)  NOT NULL,
    quota_limit    BIGINT,
    quota_used     BIGINT       NOT NULL DEFAULT 0,
    allowed_models JSONB        NOT NULL DEFAULT '[]'::jsonb,
    allowed_ips    JSONB        NOT NULL DEFAULT '[]'::jsonb,
    rpm_limit      INT          NOT NULL DEFAULT 0,
    tpm_limit      INT          NOT NULL DEFAULT 0,
    expires_at     TIMESTAMPTZ,
    last_used_at   TIMESTAMPTZ,
    status         SMALLINT     NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL,
    updated_at     TIMESTAMPTZ  NOT NULL,
    deleted_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_api_tokens_key_hash ON api_tokens (key_hash);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user_status ON api_tokens (user_id, status);
CREATE INDEX IF NOT EXISTS idx_api_tokens_expires ON api_tokens (expires_at);
CREATE INDEX IF NOT EXISTS idx_api_tokens_deleted ON api_tokens (deleted_at);

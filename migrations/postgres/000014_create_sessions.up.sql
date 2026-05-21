CREATE TABLE IF NOT EXISTS sessions (
    id             VARCHAR(48)  NOT NULL PRIMARY KEY,
    user_id        BIGINT       NOT NULL,
    ip             VARCHAR(45)  NOT NULL DEFAULT '',
    user_agent     VARCHAR(256) NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ  NOT NULL,
    last_seen_at   TIMESTAMPTZ  NOT NULL,
    expires_at     TIMESTAMPTZ  NOT NULL,
    revoked_at     TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_expires ON sessions (user_id, expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions (expires_at);

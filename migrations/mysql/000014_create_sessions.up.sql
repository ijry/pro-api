CREATE TABLE IF NOT EXISTS sessions (
    id             VARCHAR(48)  NOT NULL,
    user_id        BIGINT       NOT NULL,
    ip             VARCHAR(45)  NOT NULL DEFAULT '',
    user_agent     VARCHAR(256) NOT NULL DEFAULT '',
    created_at     DATETIME(3)  NOT NULL,
    last_seen_at   DATETIME(3)  NOT NULL,
    expires_at     DATETIME(3)  NOT NULL,
    revoked_at     DATETIME(3)  NULL,
    PRIMARY KEY (id),
    KEY idx_sessions_user_expires (user_id, expires_at),
    KEY idx_sessions_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

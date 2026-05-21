CREATE TABLE IF NOT EXISTS user_groups (
    id           BIGINT        NOT NULL PRIMARY KEY,
    name         VARCHAR(64)   NOT NULL UNIQUE,
    display_name VARCHAR(64)   NOT NULL DEFAULT '',
    ratio        NUMERIC(10,4) NOT NULL DEFAULT 1.0000,
    priority     SMALLINT      NOT NULL DEFAULT 0,
    status       SMALLINT      NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ   NOT NULL,
    updated_at   TIMESTAMPTZ   NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_user_groups_status ON user_groups (status);

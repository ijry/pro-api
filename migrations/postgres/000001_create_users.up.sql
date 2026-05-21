CREATE TABLE IF NOT EXISTS users (
    id              BIGINT       NOT NULL PRIMARY KEY,
    username        VARCHAR(64)  NOT NULL UNIQUE,
    email           VARCHAR(128) UNIQUE,
    password_hash   VARCHAR(128),
    display_name    VARCHAR(64),
    avatar          VARCHAR(256),
    role            SMALLINT     NOT NULL DEFAULT 0,
    status          SMALLINT     NOT NULL DEFAULT 0,
    group_id        BIGINT,
    primary_dept_id BIGINT,
    invited_by      BIGINT,
    last_login_at   TIMESTAMPTZ,
    last_login_ip   VARCHAR(45),
    created_at      TIMESTAMPTZ  NOT NULL,
    updated_at      TIMESTAMPTZ  NOT NULL,
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_group_status ON users (group_id, status);
CREATE INDEX IF NOT EXISTS idx_users_invited_by ON users (invited_by);

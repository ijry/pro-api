CREATE TABLE IF NOT EXISTS error_logs (
    id          BIGINT       NOT NULL PRIMARY KEY,
    created_at  TIMESTAMPTZ  NOT NULL,
    user_id     BIGINT,
    token_id    BIGINT,
    channel_id  BIGINT,
    error_code  INT          NOT NULL,
    error_type  VARCHAR(64)  NOT NULL DEFAULT '',
    stack       TEXT,
    context     JSONB,
    trace_id    VARCHAR(64)  NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_err_time      ON error_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_err_code_time ON error_logs (error_code, created_at);
CREATE INDEX IF NOT EXISTS idx_err_user_time ON error_logs (user_id, created_at);

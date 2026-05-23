CREATE TABLE IF NOT EXISTS error_logs (
    id          BIGINT       NOT NULL,
    created_at  DATETIME(3)  NOT NULL,
    user_id     BIGINT,
    token_id    BIGINT,
    channel_id  BIGINT,
    error_code  INT          NOT NULL,
    error_type  VARCHAR(64)  NOT NULL DEFAULT '',
    stack       TEXT,
    context     JSON,
    trace_id    VARCHAR(64)  NOT NULL DEFAULT '',
    PRIMARY KEY (id),
    KEY idx_err_time      (created_at),
    KEY idx_err_code_time (error_code, created_at),
    KEY idx_err_user_time (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS request_logs (
    id                    BIGINT       NOT NULL,
    created_at            TIMESTAMPTZ  NOT NULL,
    user_id               BIGINT       NOT NULL,
    token_id              BIGINT       NOT NULL,
    dept_id               BIGINT,
    group_id              BIGINT,
    event_type            SMALLINT     NOT NULL DEFAULT 0,
    client_model          VARCHAR(128) NOT NULL,
    upstream_model        VARCHAR(128) NOT NULL DEFAULT '',
    channel_id            BIGINT,
    protocol              VARCHAR(16)  NOT NULL DEFAULT 'openai',
    endpoint              VARCHAR(64)  NOT NULL,
    ip                    VARCHAR(45)  NOT NULL DEFAULT '',
    user_agent            VARCHAR(256) NOT NULL DEFAULT '',
    status_code           SMALLINT     NOT NULL DEFAULT 0,
    latency_ms            INT          NOT NULL DEFAULT 0,
    ttft_ms               INT          NOT NULL DEFAULT 0,
    stream                BOOLEAN      NOT NULL DEFAULT FALSE,
    input_tokens          INT          NOT NULL DEFAULT 0,
    output_tokens         INT          NOT NULL DEFAULT 0,
    cached_tokens         INT          NOT NULL DEFAULT 0,
    reasoning_tokens      INT          NOT NULL DEFAULT 0,
    total_quota           BIGINT       NOT NULL DEFAULT 0,
    billing_input_ratio   NUMERIC(10,4) NOT NULL DEFAULT 0,
    billing_output_ratio  NUMERIC(10,4) NOT NULL DEFAULT 0,
    billing_group_ratio   NUMERIC(10,4) NOT NULL DEFAULT 1,
    error_code            INT          NOT NULL DEFAULT 0,
    error_msg             VARCHAR(512) NOT NULL DEFAULT '',
    trace_id              VARCHAR(64)  NOT NULL DEFAULT '',
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX IF NOT EXISTS idx_req_user_time        ON request_logs (user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_req_token_time       ON request_logs (token_id, created_at);
CREATE INDEX IF NOT EXISTS idx_req_channel_time     ON request_logs (channel_id, created_at);
CREATE INDEX IF NOT EXISTS idx_req_clientmodel_time ON request_logs (client_model, created_at);
CREATE INDEX IF NOT EXISTS idx_req_event_time       ON request_logs (event_type, created_at);

-- Default partition to catch all data when specific partitions haven't been created yet
CREATE TABLE IF NOT EXISTS request_logs_default
    PARTITION OF request_logs DEFAULT;

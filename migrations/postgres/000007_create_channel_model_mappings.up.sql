CREATE TABLE IF NOT EXISTS channel_model_mappings (
    id              BIGINT         NOT NULL PRIMARY KEY,
    channel_id      BIGINT         NOT NULL,
    client_model    VARCHAR(128)   NOT NULL,
    upstream_model  VARCHAR(128)   NOT NULL,
    input_ratio     NUMERIC(10,4)  NULL,
    output_ratio    NUMERIC(10,4)  NULL,
    cached_ratio    NUMERIC(10,4)  NULL,
    reasoning_ratio NUMERIC(10,4)  NULL,
    created_at      TIMESTAMPTZ    NOT NULL,
    updated_at      TIMESTAMPTZ    NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_chm_channel_client ON channel_model_mappings (channel_id, client_model);
CREATE INDEX IF NOT EXISTS idx_chm_client_model ON channel_model_mappings (client_model);

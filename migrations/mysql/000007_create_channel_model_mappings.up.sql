CREATE TABLE IF NOT EXISTS channel_model_mappings (
    id              BIGINT         NOT NULL,
    channel_id      BIGINT         NOT NULL,
    client_model    VARCHAR(128)   NOT NULL,
    upstream_model  VARCHAR(128)   NOT NULL,
    input_ratio     DECIMAL(10,4)  NULL,
    output_ratio    DECIMAL(10,4)  NULL,
    cached_ratio    DECIMAL(10,4)  NULL,
    reasoning_ratio DECIMAL(10,4)  NULL,
    created_at      DATETIME(3)    NOT NULL,
    updated_at      DATETIME(3)    NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_chm_channel_client (channel_id, client_model),
    KEY idx_chm_client_model (client_model)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

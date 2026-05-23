CREATE TABLE IF NOT EXISTS pricing_rules (
    id              BIGINT         NOT NULL,
    scope           VARCHAR(16)    NOT NULL,
    group_id        BIGINT         NULL,
    model           VARCHAR(128)   NULL,
    input_ratio     DECIMAL(10,4)  NULL,
    output_ratio    DECIMAL(10,4)  NULL,
    cached_ratio    DECIMAL(10,4)  NULL,
    reasoning_ratio DECIMAL(10,4)  NULL,
    priority        SMALLINT       NOT NULL DEFAULT 100,
    status          TINYINT        NOT NULL DEFAULT 0,
    created_at      DATETIME(3)    NOT NULL,
    updated_at      DATETIME(3)    NOT NULL,
    PRIMARY KEY (id),
    KEY idx_pricing_lookup (scope, group_id, model, status),
    KEY idx_pricing_priority (priority, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

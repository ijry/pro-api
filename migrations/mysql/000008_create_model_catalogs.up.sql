CREATE TABLE IF NOT EXISTS model_catalogs (
    id                       BIGINT        NOT NULL,
    name                     VARCHAR(128)  NOT NULL,
    family                   VARCHAR(32)   NOT NULL DEFAULT 'chat',
    capabilities             JSON          NOT NULL,
    default_input_ratio      DECIMAL(10,4) NOT NULL DEFAULT 1.0000,
    default_output_ratio     DECIMAL(10,4) NOT NULL DEFAULT 1.0000,
    default_cached_ratio     DECIMAL(10,4) NULL,
    default_reasoning_ratio  DECIMAL(10,4) NULL,
    max_input_tokens         INT           NOT NULL DEFAULT 0,
    status                   TINYINT       NOT NULL DEFAULT 0,
    owned_by                 VARCHAR(32)   NOT NULL DEFAULT '',
    created_at               DATETIME(3)   NOT NULL,
    updated_at               DATETIME(3)   NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_model_catalogs_name (name),
    KEY idx_model_catalogs_family_status (family, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS system_settings (
    `key`       VARCHAR(128) NOT NULL,
    `value`     JSON         NOT NULL,
    description VARCHAR(256) NOT NULL DEFAULT '',
    updated_by  BIGINT,
    updated_at  DATETIME(3)  NOT NULL,
    PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

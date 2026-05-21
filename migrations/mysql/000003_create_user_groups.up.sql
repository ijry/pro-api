CREATE TABLE IF NOT EXISTS user_groups (
    id            BIGINT       NOT NULL,
    name          VARCHAR(64)  NOT NULL,
    display_name  VARCHAR(64)  NOT NULL DEFAULT '',
    ratio         DECIMAL(10,4) NOT NULL DEFAULT 1.0000,
    priority      SMALLINT     NOT NULL DEFAULT 0,
    status        TINYINT      NOT NULL DEFAULT 0,
    created_at    DATETIME(3)  NOT NULL,
    updated_at    DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_groups_name (name),
    KEY idx_user_groups_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

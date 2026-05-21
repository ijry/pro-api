CREATE TABLE IF NOT EXISTS notices (
    id          BIGINT       NOT NULL,
    title       VARCHAR(128) NOT NULL,
    content     TEXT         NOT NULL,
    level       VARCHAR(16)  NOT NULL DEFAULT 'info',
    target      VARCHAR(16)  NOT NULL DEFAULT 'all',
    status      TINYINT      NOT NULL DEFAULT 0,
    publish_at  DATETIME(3),
    expires_at  DATETIME(3),
    pinned      TINYINT(1)   NOT NULL DEFAULT 0,
    created_by  BIGINT       NOT NULL,
    created_at  DATETIME(3)  NOT NULL,
    updated_at  DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    KEY idx_notice_status_publish (status, publish_at),
    KEY idx_notice_target_status (target, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

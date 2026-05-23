CREATE TABLE IF NOT EXISTS channels (
    id          BIGINT       NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    provider    VARCHAR(32)  NOT NULL,
    base_url    VARCHAR(256) NOT NULL DEFAULT '',
    credentials TEXT         NOT NULL,
    priority    SMALLINT     NOT NULL DEFAULT 0,
    weight      INT          NOT NULL DEFAULT 1,
    status      TINYINT      NOT NULL DEFAULT 0,
    tags        JSON         NOT NULL,
    extra       JSON         NOT NULL,
    created_at  DATETIME(3)  NOT NULL,
    updated_at  DATETIME(3)  NOT NULL,
    deleted_at  DATETIME(3)  NULL,
    PRIMARY KEY (id),
    KEY idx_channel_provider_status (provider, status),
    KEY idx_channel_priority_weight (priority, weight),
    KEY idx_channel_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

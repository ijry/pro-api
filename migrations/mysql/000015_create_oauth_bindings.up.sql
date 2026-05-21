CREATE TABLE IF NOT EXISTS oauth_bindings (
    id            BIGINT       NOT NULL,
    user_id       BIGINT       NOT NULL,
    provider      VARCHAR(16)  NOT NULL,
    provider_uid  VARCHAR(128) NOT NULL,
    email         VARCHAR(128) NOT NULL DEFAULT '',
    profile       JSON         NULL,
    created_at    DATETIME(3)  NOT NULL,
    updated_at    DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_oauth_provider_uid (provider, provider_uid),
    KEY idx_oauth_user_provider (user_id, provider)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

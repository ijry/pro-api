CREATE TABLE IF NOT EXISTS wallets (
    id                      BIGINT       NOT NULL,
    owner_type              VARCHAR(16)  NOT NULL,
    owner_id                BIGINT       NOT NULL,
    quota_balance           BIGINT       NOT NULL DEFAULT 0,
    quota_total_recharged   BIGINT       NOT NULL DEFAULT 0,
    quota_total_consumed    BIGINT       NOT NULL DEFAULT 0,
    currency                VARCHAR(8)   NOT NULL DEFAULT 'USD',
    version                 INT          NOT NULL DEFAULT 0,
    created_at              DATETIME(3)  NOT NULL,
    updated_at              DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_wallets_owner (owner_type, owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

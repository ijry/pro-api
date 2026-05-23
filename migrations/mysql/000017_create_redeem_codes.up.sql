CREATE TABLE IF NOT EXISTS redeem_codes (
    id            BIGINT       NOT NULL,
    code_hash     VARCHAR(64)  NOT NULL,
    code_prefix   VARCHAR(8)   NOT NULL,
    amount_quota  BIGINT       NOT NULL,
    batch_no      VARCHAR(32)  NOT NULL DEFAULT '',
    status        TINYINT      NOT NULL DEFAULT 0,
    used_by       BIGINT       NULL,
    used_at       DATETIME(3)  NULL,
    expires_at    DATETIME(3)  NULL,
    created_by    BIGINT       NOT NULL,
    created_at    DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_redeem_codes_hash (code_hash),
    KEY idx_redeem_codes_status_expires (status, expires_at),
    KEY idx_redeem_codes_batch (batch_no),
    KEY idx_redeem_codes_used_by (used_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

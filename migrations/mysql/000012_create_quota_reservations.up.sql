CREATE TABLE IF NOT EXISTS quota_reservations (
    id                CHAR(36)     NOT NULL,
    wallet_id         BIGINT       NOT NULL,
    user_id           BIGINT       NOT NULL,
    token_id          BIGINT       NOT NULL,
    request_id        BIGINT       NULL,
    reserved_quota    BIGINT       NOT NULL,
    committed_quota   BIGINT       NOT NULL DEFAULT 0,
    status            TINYINT      NOT NULL DEFAULT 0,
    created_at        DATETIME(3)  NOT NULL,
    committed_at      DATETIME(3)  NULL,
    expires_at        DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    KEY idx_qr_status_expires (status, expires_at),
    KEY idx_qr_request (request_id),
    KEY idx_qr_wallet_created (wallet_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

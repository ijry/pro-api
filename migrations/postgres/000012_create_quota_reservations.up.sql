CREATE TABLE IF NOT EXISTS quota_reservations (
    id              CHAR(36)     NOT NULL PRIMARY KEY,
    wallet_id       BIGINT       NOT NULL,
    user_id         BIGINT       NOT NULL,
    token_id        BIGINT       NOT NULL,
    request_id      BIGINT,
    reserved_quota  BIGINT       NOT NULL,
    committed_quota BIGINT       NOT NULL DEFAULT 0,
    status          SMALLINT     NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL,
    committed_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_qr_status_expires ON quota_reservations (status, expires_at);
CREATE INDEX IF NOT EXISTS idx_qr_request ON quota_reservations (request_id);
CREATE INDEX IF NOT EXISTS idx_qr_wallet_created ON quota_reservations (wallet_id, created_at);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id              BIGINT        NOT NULL,
    wallet_id       BIGINT        NOT NULL,
    direction       VARCHAR(8)    NOT NULL,
    amount_quota    BIGINT        NOT NULL,
    amount_money    BIGINT        NOT NULL DEFAULT 0,
    currency        VARCHAR(8)    NOT NULL DEFAULT 'USD',
    ref_type        VARCHAR(16)   NOT NULL,
    ref_id          BIGINT        NULL,
    balance_after   BIGINT        NOT NULL,
    description     VARCHAR(256)  NOT NULL DEFAULT '',
    created_at      DATETIME(3)   NOT NULL,
    PRIMARY KEY (id),
    KEY idx_ledger_wallet_created (wallet_id, created_at DESC),
    KEY idx_ledger_ref (ref_type, ref_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

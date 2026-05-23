CREATE TABLE IF NOT EXISTS ledger_entries (
    id            BIGINT       NOT NULL PRIMARY KEY,
    wallet_id     BIGINT       NOT NULL,
    direction     VARCHAR(8)   NOT NULL,
    amount_quota  BIGINT       NOT NULL,
    amount_money  BIGINT       NOT NULL DEFAULT 0,
    currency      VARCHAR(8)   NOT NULL DEFAULT 'USD',
    ref_type      VARCHAR(16)  NOT NULL,
    ref_id        BIGINT,
    balance_after BIGINT       NOT NULL,
    description   VARCHAR(256) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ledger_wallet_created ON ledger_entries (wallet_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ledger_ref ON ledger_entries (ref_type, ref_id);

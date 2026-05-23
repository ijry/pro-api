CREATE TABLE IF NOT EXISTS wallets (
    id                    BIGINT       NOT NULL PRIMARY KEY,
    owner_type            VARCHAR(16)  NOT NULL,
    owner_id              BIGINT       NOT NULL,
    quota_balance         BIGINT       NOT NULL DEFAULT 0,
    quota_total_recharged BIGINT       NOT NULL DEFAULT 0,
    quota_total_consumed  BIGINT       NOT NULL DEFAULT 0,
    currency              VARCHAR(8)   NOT NULL DEFAULT 'USD',
    version               INT          NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ  NOT NULL,
    updated_at            TIMESTAMPTZ  NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_wallets_owner ON wallets (owner_type, owner_id);

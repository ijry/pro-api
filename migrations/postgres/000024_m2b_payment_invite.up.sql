ALTER TABLE users ADD COLUMN IF NOT EXISTS invited_by BIGINT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_code VARCHAR(32) UNIQUE;

CREATE TABLE IF NOT EXISTS payment_orders (
    id                BIGINT        NOT NULL PRIMARY KEY,
    user_id           BIGINT        NOT NULL,
    provider          VARCHAR(32)   NOT NULL,
    out_trade_no      VARCHAR(64)   NOT NULL UNIQUE,
    provider_order_id VARCHAR(128)  NOT NULL DEFAULT '',
    amount_cents      BIGINT        NOT NULL,
    currency          VARCHAR(8)    NOT NULL DEFAULT 'CNY',
    status            VARCHAR(16)   NOT NULL DEFAULT 'pending',
    credits           BIGINT        NOT NULL DEFAULT 0,
    meta              JSONB,
    paid_at           TIMESTAMPTZ,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status  ON payment_orders(status);

CREATE TABLE IF NOT EXISTS invite_records (
    id             BIGINT       NOT NULL PRIMARY KEY,
    inviter_id     BIGINT       NOT NULL,
    invitee_id     BIGINT       NOT NULL,
    order_id       BIGINT       NOT NULL,
    rebate_cents   BIGINT       NOT NULL DEFAULT 0,
    rebate_credits BIGINT       NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_invite_records_inviter ON invite_records(inviter_id);
CREATE INDEX IF NOT EXISTS idx_invite_records_invitee ON invite_records(invitee_id);

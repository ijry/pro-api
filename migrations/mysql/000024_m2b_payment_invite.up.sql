ALTER TABLE users ADD COLUMN invited_by BIGINT NOT NULL DEFAULT 0 COMMENT '邀请人 user_id';

CREATE TABLE IF NOT EXISTS payment_orders (
    id                BIGINT        NOT NULL PRIMARY KEY,
    user_id           BIGINT        NOT NULL,
    provider          VARCHAR(32)   NOT NULL COMMENT 'stripe/alipay/wechatpay',
    out_trade_no      VARCHAR(64)   NOT NULL UNIQUE,
    provider_order_id VARCHAR(128)  NOT NULL DEFAULT '',
    amount_cents      BIGINT        NOT NULL COMMENT '分',
    currency          VARCHAR(8)    NOT NULL DEFAULT 'CNY',
    status            VARCHAR(16)   NOT NULL DEFAULT 'pending' COMMENT 'pending/paid/failed/refunded',
    credits           BIGINT        NOT NULL DEFAULT 0,
    meta              JSON,
    paid_at           DATETIME(3),
    created_at        DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at        DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_payment_orders_user_id (user_id),
    INDEX idx_payment_orders_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS invite_records (
    id             BIGINT       NOT NULL PRIMARY KEY,
    inviter_id     BIGINT       NOT NULL,
    invitee_id     BIGINT       NOT NULL,
    order_id       BIGINT       NOT NULL,
    rebate_cents   BIGINT       NOT NULL DEFAULT 0,
    rebate_credits BIGINT       NOT NULL DEFAULT 0,
    created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_invite_records_inviter (inviter_id),
    INDEX idx_invite_records_invitee (invitee_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

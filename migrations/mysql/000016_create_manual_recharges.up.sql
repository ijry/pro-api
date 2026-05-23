CREATE TABLE IF NOT EXISTS manual_recharges (
    id              BIGINT       NOT NULL,
    user_id         BIGINT       NOT NULL,
    amount_money    BIGINT       NOT NULL,
    currency        VARCHAR(8)   NOT NULL DEFAULT 'CNY',
    amount_quota    BIGINT       NOT NULL DEFAULT 0,
    status          TINYINT      NOT NULL DEFAULT 0,
    applicant_note  VARCHAR(512) NOT NULL DEFAULT '',
    reviewer_id     BIGINT       NULL,
    review_note     VARCHAR(512) NOT NULL DEFAULT '',
    reviewed_at     DATETIME(3)  NULL,
    created_at      DATETIME(3)  NOT NULL,
    updated_at      DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    KEY idx_manual_recharges_user_status (user_id, status),
    KEY idx_manual_recharges_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

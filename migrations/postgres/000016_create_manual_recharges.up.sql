CREATE TABLE IF NOT EXISTS manual_recharges (
    id             BIGINT       NOT NULL PRIMARY KEY,
    user_id        BIGINT       NOT NULL,
    amount_money   BIGINT       NOT NULL,
    currency       VARCHAR(8)   NOT NULL DEFAULT 'CNY',
    amount_quota   BIGINT       NOT NULL DEFAULT 0,
    status         SMALLINT     NOT NULL DEFAULT 0,
    applicant_note VARCHAR(512) NOT NULL DEFAULT '',
    reviewer_id    BIGINT,
    review_note    VARCHAR(512) NOT NULL DEFAULT '',
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL,
    updated_at     TIMESTAMPTZ  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_manual_recharges_user_status    ON manual_recharges (user_id, status);
CREATE INDEX IF NOT EXISTS idx_manual_recharges_status_created ON manual_recharges (status, created_at);

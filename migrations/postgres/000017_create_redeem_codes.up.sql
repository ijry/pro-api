CREATE TABLE IF NOT EXISTS redeem_codes (
    id           BIGINT      NOT NULL PRIMARY KEY,
    code_hash    VARCHAR(64) NOT NULL,
    code_prefix  VARCHAR(8)  NOT NULL,
    amount_quota BIGINT      NOT NULL,
    batch_no     VARCHAR(32) NOT NULL DEFAULT '',
    status       SMALLINT    NOT NULL DEFAULT 0,
    used_by      BIGINT,
    used_at      TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_by   BIGINT      NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_redeem_codes_hash         ON redeem_codes (code_hash);
CREATE INDEX        IF NOT EXISTS idx_redeem_codes_status_exp  ON redeem_codes (status, expires_at);
CREATE INDEX        IF NOT EXISTS idx_redeem_codes_batch       ON redeem_codes (batch_no);
CREATE INDEX        IF NOT EXISTS idx_redeem_codes_used_by     ON redeem_codes (used_by);

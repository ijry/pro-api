CREATE TABLE IF NOT EXISTS account_events (
    id          BIGINT       NOT NULL PRIMARY KEY,
    account_id  BIGINT       NOT NULL,
    event_type  VARCHAR(32)  NOT NULL,
    payload     JSONB,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_account_events_acc_created ON account_events(account_id, created_at);
CREATE INDEX IF NOT EXISTS idx_account_events_type_created ON account_events(event_type, created_at);

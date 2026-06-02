CREATE TABLE IF NOT EXISTS account_events (
    id          BIGINT       NOT NULL PRIMARY KEY,
    account_id  BIGINT       NOT NULL,
    event_type  VARCHAR(32)  NOT NULL,
    payload     JSON         NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_account_events_acc_created (account_id, created_at),
    INDEX idx_account_events_type_created (event_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

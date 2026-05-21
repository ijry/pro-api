CREATE TABLE IF NOT EXISTS audit_logs (
    id           BIGINT       NOT NULL,
    created_at   DATETIME(3)  NOT NULL,
    actor_id     BIGINT,
    actor_role   TINYINT      NOT NULL DEFAULT 0,
    action       VARCHAR(64)  NOT NULL,
    target_type  VARCHAR(32)  NOT NULL,
    target_id    BIGINT,
    `before`     JSON,
    `after`      JSON,
    ip           VARCHAR(45)  NOT NULL DEFAULT '',
    PRIMARY KEY (id),
    KEY idx_audit_actor (actor_id, created_at),
    KEY idx_audit_target (target_type, target_id, created_at),
    KEY idx_audit_action (action, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

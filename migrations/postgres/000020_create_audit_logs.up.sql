CREATE TABLE IF NOT EXISTS audit_logs (
    id          BIGINT       NOT NULL PRIMARY KEY,
    created_at  TIMESTAMPTZ  NOT NULL,
    actor_id    BIGINT,
    actor_role  SMALLINT     NOT NULL DEFAULT 0,
    action      VARCHAR(64)  NOT NULL,
    target_type VARCHAR(32)  NOT NULL,
    target_id   BIGINT,
    before      JSONB,
    after       JSONB,
    ip          VARCHAR(45)  NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs (actor_id, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_target ON audit_logs (target_type, target_id, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs (action, created_at);

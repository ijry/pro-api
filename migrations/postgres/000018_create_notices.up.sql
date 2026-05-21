CREATE TABLE IF NOT EXISTS notices (
    id         BIGINT       NOT NULL PRIMARY KEY,
    title      VARCHAR(128) NOT NULL,
    content    TEXT         NOT NULL,
    level      VARCHAR(16)  NOT NULL DEFAULT 'info',
    target     VARCHAR(16)  NOT NULL DEFAULT 'all',
    status     SMALLINT     NOT NULL DEFAULT 0,
    publish_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    pinned     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by BIGINT       NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notice_status_publish ON notices (status, publish_at);
CREATE INDEX IF NOT EXISTS idx_notice_target_status ON notices (target, status);

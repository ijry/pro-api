CREATE TABLE IF NOT EXISTS channels (
    id          BIGINT       NOT NULL PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL,
    provider    VARCHAR(32)  NOT NULL,
    base_url    VARCHAR(256) NOT NULL DEFAULT '',
    credentials TEXT         NOT NULL,
    priority    SMALLINT     NOT NULL DEFAULT 0,
    weight      INT          NOT NULL DEFAULT 1,
    status      SMALLINT     NOT NULL DEFAULT 0,
    tags        JSONB        NOT NULL,
    extra       JSONB        NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL,
    deleted_at  TIMESTAMPTZ  NULL
);
CREATE INDEX IF NOT EXISTS idx_channel_provider_status ON channels (provider, status);
CREATE INDEX IF NOT EXISTS idx_channel_priority_weight ON channels (priority, weight);
CREATE INDEX IF NOT EXISTS idx_channel_deleted ON channels (deleted_at);

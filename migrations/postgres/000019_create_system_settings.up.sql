CREATE TABLE IF NOT EXISTS system_settings (
    key         VARCHAR(128) PRIMARY KEY,
    value       JSONB        NOT NULL,
    description VARCHAR(256) NOT NULL DEFAULT '',
    updated_by  BIGINT,
    updated_at  TIMESTAMPTZ  NOT NULL
);

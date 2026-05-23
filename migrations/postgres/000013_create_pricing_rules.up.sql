CREATE TABLE IF NOT EXISTS pricing_rules (
    id              BIGINT         NOT NULL PRIMARY KEY,
    scope           VARCHAR(16)    NOT NULL,
    group_id        BIGINT,
    model           VARCHAR(128),
    input_ratio     NUMERIC(10,4),
    output_ratio    NUMERIC(10,4),
    cached_ratio    NUMERIC(10,4),
    reasoning_ratio NUMERIC(10,4),
    priority        SMALLINT       NOT NULL DEFAULT 100,
    status          SMALLINT       NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ    NOT NULL,
    updated_at      TIMESTAMPTZ    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_pricing_lookup ON pricing_rules (scope, group_id, model, status);
CREATE INDEX IF NOT EXISTS idx_pricing_priority ON pricing_rules (priority, id);

CREATE TABLE IF NOT EXISTS model_catalogs (
    id                      BIGINT        NOT NULL PRIMARY KEY,
    name                    VARCHAR(128)  NOT NULL,
    family                  VARCHAR(32)   NOT NULL DEFAULT 'chat',
    capabilities            JSONB         NOT NULL DEFAULT '[]'::jsonb,
    default_input_ratio     NUMERIC(10,4) NOT NULL DEFAULT 1.0,
    default_output_ratio    NUMERIC(10,4) NOT NULL DEFAULT 1.0,
    default_cached_ratio    NUMERIC(10,4),
    default_reasoning_ratio NUMERIC(10,4),
    max_input_tokens        INT           NOT NULL DEFAULT 0,
    status                  SMALLINT      NOT NULL DEFAULT 0,
    owned_by                VARCHAR(32)   NOT NULL DEFAULT '',
    created_at              TIMESTAMPTZ   NOT NULL,
    updated_at              TIMESTAMPTZ   NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_model_catalogs_name ON model_catalogs (name);
CREATE INDEX IF NOT EXISTS idx_model_catalogs_family_status ON model_catalogs (family, status);

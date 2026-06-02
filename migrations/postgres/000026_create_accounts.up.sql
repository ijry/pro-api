CREATE TABLE IF NOT EXISTS accounts (
    id                       BIGINT       NOT NULL PRIMARY KEY,
    channel_id               BIGINT       NOT NULL,
    share_tag                VARCHAR(64),
    name                     VARCHAR(128) NOT NULL DEFAULT '',
    provider                 VARCHAR(32)  NOT NULL,
    tier                     VARCHAR(32)  NOT NULL DEFAULT 'unknown',
    cred_type                VARCHAR(16)  NOT NULL,
    email                    VARCHAR(128),
    external_account_id      VARCHAR(64),
    credentials              TEXT         NOT NULL,
    priority                 SMALLINT     NOT NULL DEFAULT 0,
    weight                   INT          NOT NULL DEFAULT 100,
    status                   SMALLINT     NOT NULL DEFAULT 0,
    cooldown_until           TIMESTAMPTZ,
    last_failure_at          TIMESTAMPTZ,
    last_failure_reason      VARCHAR(256) NOT NULL DEFAULT '',
    consec_failures          INT          NOT NULL DEFAULT 0,
    last_success_at          TIMESTAMPTZ,
    last_used_at             TIMESTAMPTZ,
    quota_5h_total           BIGINT,
    quota_5h_remaining       BIGINT,
    quota_5h_reset_at        TIMESTAMPTZ,
    quota_week_total         BIGINT,
    quota_week_remaining     BIGINT,
    quota_week_reset_at      TIMESTAMPTZ,
    quota_synced_at          TIMESTAMPTZ,
    access_token_expires_at  TIMESTAMPTZ,
    refresh_token_valid      SMALLINT     NOT NULL DEFAULT 0,
    last_refreshed_at        TIMESTAMPTZ,
    import_source            VARCHAR(32)  NOT NULL DEFAULT '',
    extra                    JSONB,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_accounts_channel_status ON accounts(channel_id, status);
CREATE INDEX IF NOT EXISTS idx_accounts_share_status ON accounts(share_tag, status);
CREATE INDEX IF NOT EXISTS idx_accounts_status_cooldown ON accounts(status, cooldown_until);
CREATE INDEX IF NOT EXISTS idx_accounts_provider_tier ON accounts(provider, tier);
CREATE INDEX IF NOT EXISTS idx_accounts_token_exp ON accounts(access_token_expires_at, status);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted ON accounts(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS uk_accounts_provider_extid
    ON accounts(provider, external_account_id)
    WHERE external_account_id IS NOT NULL;

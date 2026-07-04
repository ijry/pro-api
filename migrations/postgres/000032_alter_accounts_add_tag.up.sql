ALTER TABLE accounts ADD COLUMN tag VARCHAR(64) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_accounts_channel_tag ON accounts(channel_id, tag);

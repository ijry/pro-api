DROP INDEX IF EXISTS idx_accounts_channel_tag;
ALTER TABLE accounts DROP COLUMN IF EXISTS tag;

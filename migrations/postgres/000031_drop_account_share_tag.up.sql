DROP INDEX IF EXISTS idx_accounts_share_status;
ALTER TABLE accounts DROP COLUMN IF EXISTS share_tag;

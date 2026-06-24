ALTER TABLE accounts ADD COLUMN IF NOT EXISTS share_tag VARCHAR(64);
CREATE INDEX IF NOT EXISTS idx_accounts_share_status ON accounts(share_tag, status);

ALTER TABLE accounts ADD COLUMN share_tag VARCHAR(64) NULL AFTER channel_id;
ALTER TABLE accounts ADD INDEX idx_accounts_share_status (share_tag, status);

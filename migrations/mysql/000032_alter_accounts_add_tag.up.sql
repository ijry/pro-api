ALTER TABLE accounts ADD COLUMN tag VARCHAR(64) NOT NULL DEFAULT '' AFTER channel_id;
ALTER TABLE accounts ADD INDEX idx_accounts_channel_tag (channel_id, tag);

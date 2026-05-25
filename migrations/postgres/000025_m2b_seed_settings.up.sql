INSERT INTO system_settings (key, value, description, updated_at) VALUES
  ('auth.google_oauth',   '{}', 'Google OAuth 配置 {client_id,client_secret}',   NOW()),
  ('auth.wechat_oauth',   '{}', '微信 OAuth 配置 {app_id,app_secret}',            NOW()),
  ('auth.feishu_oauth',   '{}', '飞书 OAuth 配置 {app_id,app_secret}',            NOW()),
  ('auth.dingtalk_oauth', '{}', '钉钉 OAuth 配置 {client_id,client_secret}',     NOW()),
  ('auth.discord_oauth',  '{}', 'Discord OAuth 配置 {client_id,client_secret}',  NOW()),
  ('payment.stripe',      '{}', 'Stripe 配置 {secret_key,webhook_secret}',       NOW()),
  ('payment.alipay',      '{}', '支付宝配置 {app_id,private_key,public_key}',    NOW()),
  ('payment.wechatpay',   '{}', '微信支付配置 {mch_id,serial_no,private_key,api_v3_key}', NOW()),
  ('invite.rebate_ratio',  '0.1',  '邀请返佣比例(0~1)',   NOW()),
  ('invite.credit_per_cent','1.0', '每分钱对应返佣 credits', NOW()),
  ('playground.default_channel_id', '0', 'Playground 默认渠道 ID', NOW())
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

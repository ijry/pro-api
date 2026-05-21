-- 默认用户分组
INSERT INTO user_groups (id, name, display_name, ratio, priority, status, created_at, updated_at) VALUES
  (1, 'default', '普通用户', 1.0000, 0, 0, NOW(), NOW()),
  (2, 'vip',     'VIP',     0.8000, 0, 0, NOW(), NOW()),
  (3, 'svip',    'SVIP',    0.6000, 0, 0, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;

-- system_settings 默认值
INSERT INTO system_settings (key, value, description, updated_at) VALUES
  ('auth.allow_register',                'true'::jsonb,               '是否开放注册',         NOW()),
  ('auth.email_verification_required',   'true'::jsonb,               '注册邮箱验证',         NOW()),
  ('auth.password.min_length',           '8'::jsonb,                  '密码最小长度',         NOW()),
  ('auth.password.require_mixed',        'false'::jsonb,              '密码是否要混合字符',   NOW()),
  ('session.ttl_days',                   '30'::jsonb,                 'Session TTL(天)',     NOW()),
  ('session.sliding',                    'true'::jsonb,               'Session 滑动延期',     NOW()),
  ('token.default_rpm_limit',            '60'::jsonb,                 '令牌默认 RPM',         NOW()),
  ('token.default_tpm_limit',            '100000'::jsonb,             '令牌默认 TPM',         NOW()),
  ('token.prefix_show_len',              '8'::jsonb,                  '令牌前缀展示长度',     NOW()),
  ('pricing.base_quota_per_dollar',      '500000'::jsonb,             '1 美元对应 quota 数', NOW()),
  ('pricing.exchange_rate_cny_per_usd',  '7.0'::jsonb,                'CNY/USD 汇率',         NOW()),
  ('channel.selector.mode',              '"priority_weight"'::jsonb,  '渠道选择模式',         NOW()),
  ('channel.selector.max_retries',       '3'::jsonb,                  '最大重试次数',         NOW()),
  ('channel.breaker.window_seconds',     '60'::jsonb,                 '熔断统计窗口',         NOW()),
  ('channel.breaker.fail_threshold',     '5'::jsonb,                  '熔断阈值',             NOW()),
  ('channel.breaker.cool_down_seconds',  '30'::jsonb,                 '熔断冷却',             NOW()),
  ('ratelimit.user_default_rpm',         '60'::jsonb,                 '用户默认 RPM',         NOW()),
  ('ratelimit.user_default_tpm',         '100000'::jsonb,             '用户默认 TPM',         NOW()),
  ('ratelimit.ip_rpm',                   '600'::jsonb,                'IP RPM',               NOW()),
  ('ratelimit.model_default_rpm',        '0'::jsonb,                  '模型默认 RPM(0=不限)',NOW()),
  ('ratelimit.model_default_tpm',        '0'::jsonb,                  '模型默认 TPM(0=不限)',NOW()),
  ('ratelimit.enabled',                  'true'::jsonb,               '全局限流开关',         NOW()),
  ('log.request_log_retain_months',      '6'::jsonb,                  '请求日志保留(月)',    NOW()),
  ('log.error_log_retain_days',          '30'::jsonb,                 '错误日志保留(天)',    NOW()),
  ('log.flush_batch_size',               '100'::jsonb,                '日志批量写大小',       NOW()),
  ('log.flush_interval_ms',              '1000'::jsonb,               '日志最大延迟(毫秒)',  NOW()),
  ('notice.show_max',                    '5'::jsonb,                  '登录后弹未读上限',     NOW()),
  ('billing.reserve_ttl_seconds',        '600'::jsonb,                '预扣 TTL(秒)',       NOW()),
  ('billing.reconcile_interval_seconds', '30'::jsonb,                 '回收任务间隔(秒)',   NOW()),
  ('manual_recharge.enabled',            'true'::jsonb,               '手动充值开关',         NOW()),
  ('manual_recharge.exchange_rate_cny_per_usd', '7.0'::jsonb,         '手动充值汇率',         NOW()),
  ('manual_recharge.min_amount_cny',     '100'::jsonb,                '最小充值(元)',       NOW()),
  ('manual_recharge.max_amount_cny',     '100000'::jsonb,             '最大充值(元)',       NOW()),
  ('redeem.default_expires_days',        '365'::jsonb,                '兑换码默认有效期(天)',NOW())
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

-- 默认用户分组
INSERT INTO user_groups (id, name, display_name, ratio, priority, status, created_at, updated_at) VALUES
  (1, 'default', '普通用户', 1.0000, 0, 0, NOW(3), NOW(3)),
  (2, 'vip',     'VIP',     0.8000, 0, 0, NOW(3), NOW(3)),
  (3, 'svip',    'SVIP',    0.6000, 0, 0, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- system_settings 默认值
INSERT INTO system_settings (`key`, `value`, description, updated_at) VALUES
  ('auth.allow_register',                'true',                '是否开放注册',         NOW(3)),
  ('auth.email_verification_required',   'true',                '注册邮箱验证',         NOW(3)),
  ('auth.password.min_length',           '8',                   '密码最小长度',         NOW(3)),
  ('auth.password.require_mixed',        'false',               '密码是否要混合字符',   NOW(3)),
  ('session.ttl_days',                   '30',                  'Session TTL(天)',     NOW(3)),
  ('session.sliding',                    'true',                'Session 滑动延期',     NOW(3)),
  ('token.default_rpm_limit',            '60',                  '令牌默认 RPM',         NOW(3)),
  ('token.default_tpm_limit',            '100000',              '令牌默认 TPM',         NOW(3)),
  ('token.prefix_show_len',              '8',                   '令牌前缀展示长度',     NOW(3)),
  ('pricing.base_quota_per_dollar',      '500000',              '1 美元对应 quota 数', NOW(3)),
  ('pricing.exchange_rate_cny_per_usd',  '7.0',                 'CNY/USD 汇率',         NOW(3)),
  ('channel.selector.mode',              '"priority_weight"',   '渠道选择模式',         NOW(3)),
  ('channel.selector.max_retries',       '3',                   '最大重试次数',         NOW(3)),
  ('channel.breaker.window_seconds',     '60',                  '熔断统计窗口',         NOW(3)),
  ('channel.breaker.fail_threshold',     '5',                   '熔断阈值',             NOW(3)),
  ('channel.breaker.cool_down_seconds',  '30',                  '熔断冷却',             NOW(3)),
  ('ratelimit.user_default_rpm',         '60',                  '用户默认 RPM',         NOW(3)),
  ('ratelimit.user_default_tpm',         '100000',              '用户默认 TPM',         NOW(3)),
  ('ratelimit.ip_rpm',                   '600',                 'IP RPM',               NOW(3)),
  ('ratelimit.model_default_rpm',        '0',                   '模型默认 RPM(0=不限)',NOW(3)),
  ('ratelimit.model_default_tpm',        '0',                   '模型默认 TPM(0=不限)',NOW(3)),
  ('ratelimit.enabled',                  'true',                '全局限流开关',         NOW(3)),
  ('log.request_log_retain_months',      '6',                   '请求日志保留(月)',    NOW(3)),
  ('log.error_log_retain_days',          '30',                  '错误日志保留(天)',    NOW(3)),
  ('log.flush_batch_size',               '100',                 '日志批量写大小',       NOW(3)),
  ('log.flush_interval_ms',              '1000',                '日志最大延迟(毫秒)',  NOW(3)),
  ('notice.show_max',                    '5',                   '登录后弹未读上限',     NOW(3)),
  ('billing.reserve_ttl_seconds',        '600',                 '预扣 TTL(秒)',       NOW(3)),
  ('billing.reconcile_interval_seconds', '30',                  '回收任务间隔(秒)',   NOW(3)),
  ('manual_recharge.enabled',            'true',                '手动充值开关',         NOW(3)),
  ('manual_recharge.exchange_rate_cny_per_usd', '7.0',          '手动充值汇率',         NOW(3)),
  ('manual_recharge.min_amount_cny',     '100',                 '最小充值(元)',       NOW(3)),
  ('manual_recharge.max_amount_cny',     '100000',              '最大充值(元)',       NOW(3)),
  ('redeem.default_expires_days',        '365',                 '兑换码默认有效期(天)',NOW(3))
ON DUPLICATE KEY UPDATE description = VALUES(description);

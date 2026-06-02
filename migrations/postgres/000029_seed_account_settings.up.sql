INSERT INTO system_settings (key, value, description, updated_at) VALUES
  ('account.refresher.advance_seconds',    '300',  '刷新 OAuth access_token 的提前秒数(过期前 N 秒触发)', NOW()),
  ('account.refresher.tick_seconds',       '60',   'Refresher 后台扫描间隔(秒)',                       NOW()),
  ('account.cooldown.default_seconds',     '60',   '默认 cooldown 时长(秒,无 Retry-After 时使用)',      NOW()),
  ('account.cooldown.rate_limit_fallback', '300',  '上游 429 但无 Retry-After 时的 cooldown 时长(秒)',  NOW()),
  ('account.consec_fail_threshold',        '5',    '连续失败次数阈值,达到后熔断为 cooldown',            NOW()),
  ('account.probe.timeout_ms',             '3000', '入池探测的单次超时(毫秒)',                          NOW()),
  ('account.probe.concurrency',            '8',    '批量探测的最大并发',                                NOW()),
  ('account.events.retain_days',           '90',   'account_events 保留天数(超过将被 reaper 清理)',     NOW())
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

DELETE FROM system_settings WHERE `key` IN (
  'account.refresher.advance_seconds',
  'account.refresher.tick_seconds',
  'account.cooldown.default_seconds',
  'account.cooldown.rate_limit_fallback',
  'account.consec_fail_threshold',
  'account.probe.timeout_ms',
  'account.probe.concurrency',
  'account.events.retain_days'
);

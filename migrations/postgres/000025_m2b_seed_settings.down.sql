DELETE FROM system_settings WHERE key IN (
  'auth.google_oauth','auth.wechat_oauth','auth.feishu_oauth','auth.dingtalk_oauth','auth.discord_oauth',
  'payment.stripe','payment.alipay','payment.wechatpay',
  'invite.rebate_ratio','invite.credit_per_cent','playground.default_channel_id'
);

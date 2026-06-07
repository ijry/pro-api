# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added — M2 MVP β SaaS(进行中)

- 入口协议互转:Anthropic `/v1/messages`、Gemini `/v1beta/models/:model/generateContent`(均含流式)入口 + IR 归一化,任意入口可路由任意上游
- 适配器追加至 18 家(M2a 新增 Groq / Mistral / 零一万物 / OpenRouter / Hugging Face / MiniMax / 腾讯混元;M2b 新增 Cohere / 讯飞星火,均走 OpenAI 兼容端点)
- 多模态:图像生成 / TTS / STT / Embeddings 接入 relay
- OAuth 六家完整(新增 Google / 微信 / 飞书 / 钉钉 / Discord)
- 在线支付:Stripe / 支付宝 / 微信支付,统一 provider 抽象
- 邀请返佣:邀请码 / 邀请记录 / 返佣汇总
- 账号池:上游账号统一纳管 + 与渠道绑定
- 计费集成:relay 挂分组倍率中间件,Biller / Pricing / Log 依赖注入接线
- 前端:用户前台 Playground / 模型广场 / 邀请页;后台账号池 / 分组 / 定价页;中英双语 i18n

> 未完成(M2 收尾):异步任务系统(asynq)+ Midjourney / Suno;账号池 OAuth 拉号(PKCE)。

### Added — M1 MVP α 核心闭环

- 用户系统:邮箱密码注册登录、邮箱验证码、GitHub OAuth、Session + Redis
- API 令牌:生成 / 吊销 / 限额 / IP 与模型白名单 / 过期
- 适配器层 9 家:OpenAI / Azure / Anthropic / Gemini / DeepSeek / Moonshot / 智谱 / 通义 / 豆包
- 渠道:CRUD + 优先级 + 权重 + 模型映射 + 熔断状态机 + 故障转移重试
- 计费:Redis Lua 预扣 / 提交 / 退款,模型倍率 + 分组倍率
- 日志:请求 / 消费明细 + 错误日志独立;审计日志
- 限流:用户 / 令牌 / IP / 模型 多维度滑动窗口(Lua)
- 公告系统;系统设置(KV 运行时热更)
- 后台前端(Naive UI)与用户前台全页面
- 支付:手动充值审核 + 兑换码

### Added — M0 工程脚手架

- Go 后端骨架(Gin / Viper / zap / GORM 双库 / Redis / Prometheus / healthz)
- 错误码体系(`pkg/apierr`)、Snowflake ID、AES-256-GCM
- 前端 monorepo(pnpm workspaces):`web/admin`(Naive UI)+ `web/user`(shadcn 风格)+ `web/shared`
- 文档站(`docs-site/`,VitePress)
- 第一份数据库迁移(users 骨架)
- Docker Compose 开发依赖
- GitHub Actions CI(lint / unit / integration / build)
- 总体路线图与 M0 实施计划

---
title: 更新日志
outline: deep
---

# 更新日志

> 完整变更见仓库根目录的 [CHANGELOG.md](https://github.com/ijry/pro-api/blob/main/CHANGELOG.md)。本页仅同步**最近 2-3 个 minor 版本**的精简版本。

## Unreleased

> 自上一个 release 以来的变更。M1 GA 时会从这里搬到对应版本节。

### Added

- **docs-site**:补全 M1 用户文档(快速上手、架构、模块、API、部署、二次开发、计费说明、FAQ)
- **adapter**:OpenAI / Anthropic / Gemini / DeepSeek / Moonshot / Zhipu / Qwen / Doubao / Azure 9 家适配器骨架
- **billing**:Reserve / Commit / Refund Lua 原子脚本 + append-only ledger
- **ratelimit**:4 维度 Redis 滑动窗口
- **channel**:priority+weight 选渠道 + 熔断状态机 + 集群广播
- **auth**:邮箱密码 / 邮箱验证码 / GitHub OAuth 注册登录
- **token**:`pa-` 令牌 + 模型/IP 白名单 + RPM/TPM
- **payment**:手动充值审批 + 兑换码批量生成与兑换
- **system_settings**:运行时配置 + Redis Pub/Sub 集群热更新
- **observability**:Prometheus `/metrics` 端点

### Changed

- **TBD**(M1 GA 时填)

### Fixed

- **TBD**

### Removed

- **TBD**

## v0.1.0(YYYY-MM-DD)

> M1 GA。本节由 plan 阶段在 M1 闭环时填充。

### Added

- 初版完整 OpenAI 协议代理
- 9 家上游适配器(出向)
- 双轨计费体系
- 4 维度限流
- GitHub OAuth
- 手动充值 + 兑换码
- admin / user 前台 Vue3 骨架
- VitePress 文档站

### Breaking changes

- 无(首个 minor 版本)

## v0.0.x

M0 阶段骨架版本,仅作为开发 milestone,**不建议生产使用**。详见仓库 git 历史。

---

## 版本号规则

pro-api 遵循 [SemVer 2.0](https://semver.org/lang/zh-CN/):

- `MAJOR.MINOR.PATCH`
- M1 期间所有版本 < 1.0,**API 可能有 breaking change**,务必看 CHANGELOG
- 1.0 计划在 M3 完成后发布,届时正式承诺 API 稳定性

升级流程见 [升级指南](./guide/upgrade.md)。

## 关键要点

- 本页只是 docs-site 上的镜像,**真相在仓库根 [CHANGELOG.md](https://github.com/ijry/pro-api/blob/main/CHANGELOG.md)**
- M2 会接入构建脚本自动从 CHANGELOG.md 同步到本页
- 0.x 阶段的 breaking change 不算 SemVer 违规,但 changelog 会标注 `Breaking:`

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added — M0 工程脚手架

- Go 后端骨架(Gin / Viper / zap / GORM 双库 / Redis / Prometheus / healthz)
- 错误码体系(`pkg/apierr`)、Snowflake ID、AES-256-GCM
- 前端 monorepo(pnpm workspaces):`web/admin`(Naive UI)+ `web/user`(shadcn 风格)+ `web/shared`
- 文档站(`docs-site/`,VitePress)
- 第一份数据库迁移(users 骨架)
- Docker Compose 开发依赖
- GitHub Actions CI(lint / unit / integration / build)
- 总体路线图与 M0 实施计划

# 贡献指南

## 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/v1.0.0/):

```
feat(adapter):       新增 anthropic 适配器
fix(billing):        修复流式 token 计数偏差
docs(api):           更新模型广场接口文档
chore(deps):         升级 go-redis
ci:                  添加 PG 17 矩阵
```

`scope` 必填,对应 `internal/` 一级目录或 `web/<工程>` 或 `docs-site`。

## 流程

1. fork 仓库 → 新建 `feat/<scope>` 分支
2. 改完后 `make lint && make test` 通过
3. 提 PR,关联到对应 spec(`docs/superpowers/specs/`)
4. 至少 1 名 maintainer approve 后 squash merge

## 代码风格

- Go:`gofumpt -l -w .` + `golangci-lint run`
- TypeScript / Vue:`pnpm -C web lint` + `pnpm -C web typecheck`

## 测试要求

- 核心包(`adapter / billing / channel / protocol / relay / ratelimit`)单元测试覆盖率 ≥ 80%
- 涉及数据访问的改动必须有 `dockertest` 集成测试
- UI 关键流程用 Playwright 覆盖

---
title: 贡献指南
outline: deep
---

# 贡献指南

> 本页是仓库根 [CONTRIBUTING.md](https://github.com/ijry/pro-api/blob/main/CONTRIBUTING.md) 的摘要 + 文档站本地化版本。完整流程以仓库根的 CONTRIBUTING.md 为准。

## 提 Issue

- **Bug report**:用 issue 模板,**必须附** `X-Request-ID` + 日志片段;复现步骤越具体越快定位
- **Feature request**:先描述场景(为什么需要),再讨论方案(怎么做)
- **安全漏洞**:**不开 issue**,通过 [SECURITY.md](https://github.com/ijry/pro-api/blob/main/SECURITY.md) 流程私报

## 提 PR

### 分支

```
main             ← 始终可发布,受保护
feat/xxx         ← 功能开发
fix/xxx          ← bug 修复
docs/xxx         ← 仅文档
chore/xxx        ← 杂项 / 构建 / 依赖升级
```

强烈推荐用 **rebase** 或 **squash** 合并,保持 main 的提交线性。

### Commit Message

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范:

```
feat(adapter): add mistral adapter
fix(billing): refund on stream interruption
docs(api): add openai-compat examples
chore: bump deps
refactor(channel): extract breaker logic
test(token): cover IP whitelist edge cases
```

scope 推荐用模块名(`adapter` / `billing` / `channel` / `auth` / `token` / `api` / `web` / `docs` / ...)。

### CI 必须过

提 PR 后 GitHub Actions 自动跑:

- `make test` —— 单测 + 集成测(dockertest 起 MySQL / PG / Redis)
- `make lint` —— `go vet` + `golangci-lint` + `eslint`
- 覆盖率不下降(对比 main)

任一失败 → PR 不能合。

### PR 描述模板

仓库根 `.github/PULL_REQUEST_TEMPLATE.md` 已经准备好,提 PR 时会自动填充。包含:

- **背景**:解决什么问题 / 实现什么需求
- **方案**:做了什么改动 / 为什么这么做
- **测试**:跑了哪些测试 / 验证了哪些场景
- **影响面**:有无 breaking change / 配置变更 / 数据库变更

## 代码规范

### Go

- `gofmt` + `golangci-lint`(配置见仓库根 `.golangci.yml`)
- 导出符号必须有 doc comment(`golint` 强制)
- **错误处理**:统一用 `pkg/apierr.Error` —— 不要直接 `fmt.Errorf` 抛业务错
- **interface 在调用方**:adapter / payment / oauth provider 等接口定义在使用方包,实现在子包

### TS / Vue

- `eslint` + `prettier`(配置见 `web/admin/.eslintrc.cjs` 等)
- Composition API + `<script setup>`
- Vue 单文件组件,**不要混 JSX**

### 注释

- 公开 API(导出符号)必须 doc comment
- 复杂逻辑(算法 / 状态机)写 `// why` 注释,**不写 `// what`**
- TODO 必须带 `TODO(name): ...`,方便后续追责

## TDD 推荐

新功能强烈推荐 **测试驱动开发**(TDD):

1. 先写测试(失败)
2. 再写最少代码让测试通过
3. 重构

具体见 superpowers 的 `test-driven-development` 指南(仓库 `docs/superpowers/` 下)。

虽然**不强制**,但 reviewer 会对没测试的 PR 提更多问题。

## 文档同步

改了代码,**对应的 docs-site 页也要跟上**:

| 改了什么 | 同步什么 |
|---|---|
| 改了 API endpoint / 字段 | `docs-site/zh/api/*` |
| 改了配置项 | `docs-site/zh/guide/configuration.md` |
| 改了 schema(`migrations/`) | 对应模块的 `docs-site/zh/modules/*` + spec |
| 新增 provider 适配器 | `docs-site/zh/architecture/adapter-layer.md` + `billing-guide/model-pricing.md` |
| 改了运维行为 | `docs-site/zh/deployment/*` 或 `guide/upgrade.md` |

PR 模板里有 "文档已同步" checkbox,reviewer 会检查。

## 行为准则

pro-api 项目采纳 [Contributor Covenant 2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/),完整文本见仓库根 [CODE_OF_CONDUCT.md](https://github.com/ijry/pro-api/blob/main/CODE_OF_CONDUCT.md)。

简而言之:**对人友好,对事专业**。

## 关键要点

- 安全漏洞**不开 issue** —— 单独走 SECURITY.md 邮箱报告
- 文档同步是**常被忽略**的一环,reviewer 应该检查 PR 是否动了对应文档
- conventional commits 是 CI 强制项(commit-msg hook),不规范的会被拒
- 没必要 panic:Go 错误用 `apierr.Error` 抛,中间件统一捕获并返 JSON

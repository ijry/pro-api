# 发布链补齐与品牌文案统一设计稿

- 日期: 2026-06-24
- 状态: 已与用户确认，等待写实施计划

## 1. 背景

当前仓库已经具备可发布的基础构件：

- `deploy/Dockerfile` 可构建包含前端与文档静态资源的单镜像。
- CI 已能执行 `go build -tags embed ./cmd/proapi`。
- 中文安装、部署、升级文档已经按“存在 Docker 镜像与 GitHub Releases 二进制”来写。

但真实状态并不一致：

- 公开文档写的是 `ghcr.io/proapi/proapi`，实际仓库没有对应发布链。
- 文档声明可从 GitHub Releases 下载预编译归档，实际没有自动生成这些产物。
- 页面与文档中的用户可见品牌名同时存在 `proapi` 与 `pro-api` 两种写法。

这导致两个问题：

- 安装文档会误导用户，复制命令后直接失败。
- 品牌呈现不统一，演示页、后台、用户前台与文档站的视觉一致性不足。

## 2. 目标

本次改动一次补齐以下能力：

1. 新增正式 release workflow，在打 `v*` tag 时自动发布 Docker 镜像与多平台二进制。
2. 将安装/部署/升级文档修正为真实可用的发布地址与产物命名。
3. 将用户可见的品牌展示统一为 `pro-api`。

## 3. 范围

### 3.1 In Scope

- 新增 `.github/workflows/release.yml`
- 发布 Docker 镜像到 `ghcr.io/ijry/pro-api`
- 发布 GitHub Release 二进制附件
- 修正文档中与安装、Docker、升级、反向代理相关的发布地址和示例命令
- 统一前端和 docs-site 中用户可见品牌文案为 `pro-api`

### 3.2 Out of Scope

- 不修改 npm package 名称，如 `@proapi/*`
- 不修改 cookie / localStorage / env key 等内部技术标识
- 不修改 Go module path、import path、仓库路径 `github.com/ijry/pro-api`
- 不引入 Docker Hub 双发
- 不实现额外的发布审批、签名、公证或 SBOM 流程

## 4. 设计决策

### 4.1 镜像仓库命名

镜像统一发布到 `ghcr.io/ijry/pro-api`。

原因：

- 与当前 GitHub 仓库 owner 对齐，最容易用 `GITHUB_TOKEN` 完成 GHCR 推送。
- 不再依赖并不存在或未确认的 `ghcr.io/proapi/proapi` namespace。
- 文档可以直接引用真实可推导出的镜像地址。

### 4.2 品牌改名边界

仅修改“用户可见文本”：

- HTML `<title>`
- 页面头部、登录页、副标题、页脚
- docs-site 首页及品牌展示文案
- mock 演示页里的静态标题
- i18n 中 `brand.name`、`brand.subtitle` 等展示字段

不修改“内部标识”：

- `proapi.locale`
- `proapi.theme`
- `proapi_csrf`
- `@proapi/shared`

这样可以保证视觉统一，同时避免不必要的兼容性风险。

## 5. 发布链设计

### 5.1 触发方式

新增 `.github/workflows/release.yml`：

- `push.tags: [\"v*\"]`
- `workflow_dispatch`

`workflow_dispatch` 主要用于手工回归或 dry-run 风格的调试；正式用户可见版本仍以 tag 为准。

### 5.2 发布前构建内容

release workflow 需要产出两类正式制品：

- Docker 镜像
- GitHub Release 二进制归档

构建原则：

- 复用现有 `deploy/Dockerfile`
- 二进制使用 `go build -tags embed`，确保 admin、user、docs-site 都被嵌入
- 版本信息通过 `-ldflags` 注入 `Version` / `Commit` / `BuildTime`

### 5.3 Docker 镜像

使用 `docker/setup-buildx-action` + `docker/build-push-action`。

平台：

- `linux/amd64`
- `linux/arm64`

镜像标签：

- `ghcr.io/ijry/pro-api:latest`
- `ghcr.io/ijry/pro-api:vX.Y.Z`
- `ghcr.io/ijry/pro-api:vX.Y`
- `ghcr.io/ijry/pro-api:vX`

只有 tag 发布时推送 `latest`；手动触发时可只做构建与 metadata 演练，避免误推无版本镜像。

### 5.4 GitHub Release 二进制

二进制采用 matrix 构建以下组合：

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`

附件命名保持与现有文档一致：

- `proapi_linux_amd64.tar.gz`
- `proapi_linux_arm64.tar.gz`
- `proapi_darwin_amd64.tar.gz`
- `proapi_darwin_arm64.tar.gz`
- `proapi_windows_amd64.zip`

每个归档中至少包含：

- `proapi` 或 `proapi.exe`
- `LICENSE`
- `README.md`

同时生成对应 checksum 文件，供文档中的校验步骤使用。

### 5.5 Release 元数据

Release 名称使用 tag 本身，如 `v0.2.0`。

正文优先复用自动生成的 release notes，避免每次手写说明成为额外阻塞点。

## 6. 文档修正设计

### 6.1 需要修正的中文页面

至少覆盖以下页面：

- `docs-site/zh/guide/installation.md`
- `docs-site/zh/deployment/docker.md`
- `docs-site/zh/deployment/docker-compose.md`
- `docs-site/zh/deployment/reverse-proxy.md`
- `docs-site/zh/guide/upgrade.md`

修正内容：

- 所有镜像地址改为 `ghcr.io/ijry/pro-api`
- 所有 release 下载地址保持 `https://github.com/ijry/pro-api/releases`
- 安装、升级和回滚命令与真实产物命名一致
- checksum 示例与 workflow 生成的文件一致

### 6.2 英文页面策略

当前英文大多仍是占位页。本次不强行补齐完整英文安装文档，但要保证：

- docs-site 首页等用户可见品牌展示同步统一
- 已存在的英文可见品牌文案不再写成 `proapi`

## 7. 前端与文案设计

### 7.1 修改对象

优先处理以下用户可见入口：

- `web/admin/index.html`
- `web/user/index.html`
- `web/admin/src/layouts/*`
- `web/user/src/components/biz/*`
- `web/shared/src/i18n/{zh,en}.json`
- `docs-site/{index.md,zh/index.md,en/index.md}`
- `docs-site/public/{admin-demo,user-demo}/index.html`

### 7.2 不改的地方

以下内容即使包含 `proapi` 也不在本次改动范围：

- 包名
- mock 邮箱、示例域名中作为内部 demo 标识的值
- 配置 key
- 代码中的变量名与目录名

如果个别字符串兼具“展示”和“内部协议”双重语义，优先保留兼容性，必要时只改展示层拼接文本。

## 8. 错误处理与兼容性

本次不引入运行时业务逻辑变更，主要风险集中在发布链：

- Docker workflow 权限不足会导致 GHCR push 失败
- 二进制打包脚本若未正确处理 Windows 后缀，会造成附件不可执行
- 文档命令与真实文件名一旦不一致，仍会继续误导用户

对应控制措施：

- workflow 显式声明 `contents: write` 与 `packages: write`
- 归档脚本按平台分支处理 `proapi.exe`
- 用 repo 内全文搜索校验 `ghcr.io/proapi/proapi` 是否已清干净

## 9. 验证方案

至少执行以下验证：

- `git diff --check`
- `go build -tags embed ./cmd/proapi`
- `pnpm -C web/admin build`
- `pnpm -C web/user build`
- `pnpm -C docs-site build`
- 对 `.github/workflows/release.yml` 做 YAML 结构检查
- 全文搜索以下关键字并确认结果符合预期：
  - `ghcr.io/proapi/proapi`
  - `proapi ·`
  - `>proapi<`

如本地不具备真实 GitHub tag/release 环境，则不做实际发布，只验证 workflow 结构与制品命名逻辑。

## 10. 验收标准

- 打 `v*` tag 后，仓库具备自动发布 Docker 镜像和 GitHub Release 二进制的 workflow
- 中文安装/部署/升级文档中的命令与目标地址真实可用
- 用户可见品牌名称统一为 `pro-api`
- 内部技术标识保持兼容，不发生无关重命名
- 不覆盖用户已存在的未提交改动

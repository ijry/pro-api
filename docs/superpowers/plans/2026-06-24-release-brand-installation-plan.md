# Release Pipeline, Installation Docs, and pro-api Branding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a real tag-driven release pipeline that publishes GHCR Docker images and GitHub Release binaries, then align installation docs and all user-visible branding with those real artifacts.

**Architecture:** Reuse the existing embedded build flow and `deploy/Dockerfile`, add one dedicated release workflow for tag builds, and keep documentation and visible copy aligned with the published artifact names and image registry. Limit branding edits to display text only so runtime compatibility keys and package/module identifiers remain untouched.

**Tech Stack:** GitHub Actions, Docker Buildx, Go 1.25, pnpm/VitePress, Vue 3, existing repo docs and i18n JSON

## Global Constraints

- Docker images must publish to `ghcr.io/ijry/pro-api`.
- Release binaries must use embedded assets via `go build -tags embed`.
- Binary artifact names must match the existing docs naming scheme exactly:
  - `proapi_linux_amd64.tar.gz`
  - `proapi_linux_arm64.tar.gz`
  - `proapi_darwin_amd64.tar.gz`
  - `proapi_darwin_arm64.tar.gz`
  - `proapi_windows_amd64.zip`
- User-visible brand text must be unified to `pro-api`.
- Do not rename internal technical identifiers such as `@proapi/*`, `proapi.locale`, `proapi.theme`, `proapi_csrf`, import paths, or the Go module path.
- Do not overwrite unrelated local changes in `web/admin/src/router/guard.ts` or `web/user/src/router/index.ts`.

---

## File Structure

- Create: `.github/workflows/release.yml`
  - Tag-triggered release workflow for Docker and binary publishing.
- Create: `scripts/package-release.sh`
  - Cross-platform packaging helper that wraps built binaries into the exact archive names expected by docs and release uploads.
- Modify: `docs-site/zh/guide/installation.md`
  - Replace nonexistent registry/image references and align install commands with real release outputs.
- Modify: `docs-site/zh/deployment/docker.md`
  - Align Docker deployment examples and tag strategy with GHCR release output.
- Modify: `docs-site/zh/deployment/docker-compose.md`
  - Align compose example image reference with GHCR release output.
- Modify: `docs-site/zh/deployment/reverse-proxy.md`
  - Align example image references used in proxy/container snippets.
- Modify: `docs-site/zh/guide/upgrade.md`
  - Align upgrade, rollback, and release download commands with actual artifacts.
- Modify: `web/admin/index.html`
  - Visible browser title brand text.
- Modify: `web/user/index.html`
  - Visible browser title brand text.
- Modify: `web/admin/src/layouts/AppLayout.vue`
  - Visible shell brand text.
- Modify: `web/admin/src/layouts/AuthLayout.vue`
  - Visible login/footer brand text.
- Modify: `web/user/src/components/biz/AppHeader.vue`
  - Visible header brand text.
- Modify: `web/user/src/components/biz/AppFooter.vue`
  - Visible footer brand text.
- Modify: `web/shared/src/i18n/zh.json`
  - Visible `brand.name`.
- Modify: `web/shared/src/i18n/en.json`
  - Visible `brand.name`.
- Modify: `docs-site/index.md`
  - Visible landing page hero name.
- Modify: `docs-site/zh/index.md`
  - Visible zh landing hero/alt brand text.
- Modify: `docs-site/en/index.md`
  - Visible en landing hero/alt brand text.
- Modify: `docs-site/public/admin-demo/index.html`
  - Visible demo title fallback.
- Modify: `docs-site/public/user-demo/index.html`
  - Visible demo title fallback.

### Task 1: Add release packaging and GitHub Actions workflow

**Files:**
- Create: `.github/workflows/release.yml`
- Create: `scripts/package-release.sh`
- Modify: `README.md`
- Test: `.github/workflows/release.yml`

**Interfaces:**
- Consumes: `deploy/Dockerfile`, `internal/version/version.go`, `cmd/proapi/main.go`
- Produces:
  - A tag-triggered workflow named `Release`
  - A packaging helper invoked as `bash scripts/package-release.sh <goos> <goarch> <version> <binary_path> <output_dir>`
  - Release archives and `.sha256` files with doc-compatible names

- [ ] **Step 1: Write the failing validation checklist for the release workflow**

```text
Expected artifacts for tag v0.2.0:
- ghcr.io/ijry/pro-api:v0.2.0
- ghcr.io/ijry/pro-api:v0.2
- ghcr.io/ijry/pro-api:v0
- ghcr.io/ijry/pro-api:latest
- GitHub Release attachments:
  - proapi_linux_amd64.tar.gz
  - proapi_linux_arm64.tar.gz
  - proapi_darwin_amd64.tar.gz
  - proapi_darwin_arm64.tar.gz
  - proapi_windows_amd64.zip
  - one .sha256 file per archive
```

- [ ] **Step 2: Verify the workflow does not exist yet**

Run: `Test-Path .github/workflows/release.yml`
Expected: `False`

- [ ] **Step 3: Create the packaging helper**

```bash
#!/usr/bin/env bash
set -euo pipefail

goos="${1:?goos required}"
goarch="${2:?goarch required}"
version="${3:?version required}"
binary_path="${4:?binary path required}"
output_dir="${5:?output dir required}"

mkdir -p "${output_dir}"

base="proapi_${goos}_${goarch}"
stage_dir="${output_dir}/${base}"
rm -rf "${stage_dir}"
mkdir -p "${stage_dir}"

binary_name="proapi"
if [ "${goos}" = "windows" ]; then
  binary_name="proapi.exe"
fi

cp "${binary_path}" "${stage_dir}/${binary_name}"
cp LICENSE "${stage_dir}/LICENSE"
cp README.md "${stage_dir}/README.md"

archive_path=""
if [ "${goos}" = "windows" ]; then
  archive_path="${output_dir}/${base}.zip"
  rm -f "${archive_path}"
  (
    cd "${stage_dir}"
    zip -q -r "${archive_path}" .
  )
else
  archive_path="${output_dir}/${base}.tar.gz"
  tar -C "${stage_dir}" -czf "${archive_path}" .
fi

if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "$(basename "${archive_path}")" > "${archive_path}.sha256"
else
  shasum -a 256 "$(basename "${archive_path}")" > "${archive_path}.sha256"
fi
```

- [ ] **Step 4: Create the release workflow**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

permissions:
  contents: write
  packages: write

env:
  GO_VERSION: "1.25"
  NODE_VERSION: "20"
  PNPM_VERSION: "9.12.0"
  REGISTRY_IMAGE: ghcr.io/ijry/pro-api

jobs:
  build-binaries:
    name: Build ${{ matrix.goos }}/${{ matrix.goarch }}
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        include:
          - goos: linux
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: darwin
            goarch: amd64
          - goos: darwin
            goarch: arm64
          - goos: windows
            goarch: amd64
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - uses: pnpm/action-setup@v4
        with:
          version: ${{ env.PNPM_VERSION }}

      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: pnpm
          cache-dependency-path: |
            web/pnpm-lock.yaml
            docs-site/pnpm-lock.yaml

      - run: pnpm -C web install --frozen-lockfile
      - run: pnpm -C web/admin build
      - run: pnpm -C web/user build
      - run: pnpm -C docs-site install --frozen-lockfile
      - run: pnpm -C docs-site build

      - name: Build embedded binary
        shell: bash
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: 0
          PROAPI_MASTER_KEY: "0123456789abcdef0123456789abcdef"
          VERSION: ${{ github.ref_name }}
          COMMIT: ${{ github.sha }}
          BUILD_TIME: ${{ github.event.head_commit.timestamp || github.run_id }}
        run: |
          mkdir -p dist/bin
          binary="dist/bin/proapi"
          if [ "${GOOS}" = "windows" ]; then
            binary="dist/bin/proapi.exe"
          fi
          go build -tags embed \
            -ldflags="-s -w -X github.com/ijry/pro-api/internal/version.Version=${VERSION} -X github.com/ijry/pro-api/internal/version.Commit=${COMMIT::7} -X github.com/ijry/pro-api/internal/version.BuildTime=${BUILD_TIME}" \
            -o "${binary}" ./cmd/proapi
          bash scripts/package-release.sh "${GOOS}" "${GOARCH}" "${VERSION}" "${binary}" dist/release

      - uses: actions/upload-artifact@v4
        with:
          name: release-${{ matrix.goos }}-${{ matrix.goarch }}
          path: dist/release/*

  docker:
    name: Publish Docker image
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: docker/setup-qemu-action@v3
      - uses: docker/setup-buildx-action@v3

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY_IMAGE }}
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            type=raw,value=latest

      - uses: docker/build-push-action@v6
        with:
          context: .
          file: deploy/Dockerfile
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          build-args: |
            VERSION=${{ github.ref_name }}
            COMMIT=${{ github.sha }}
            BUILD_TIME=${{ github.event.head_commit.timestamp || github.run_id }}

  release:
    name: Publish GitHub Release
    runs-on: ubuntu-latest
    needs:
      - build-binaries
      - docker
    steps:
      - uses: actions/download-artifact@v4
        with:
          path: dist/artifacts

      - name: Flatten artifacts
        shell: bash
        run: |
          mkdir -p dist/release
          find dist/artifacts -type f -exec cp {} dist/release/ \;

      - uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
          files: dist/release/*
```

- [ ] **Step 5: Update README status to mention real release artifacts**

```md
## 5-minute Quickstart

### Prerequisites

- Go 1.22+
- Node.js 20+ and pnpm 9
- Docker (for MySQL / PostgreSQL / Redis). Podman users: see `README_zh.md` for the socket adapter snippet.

For tagged releases, this repo also publishes:

- Docker images at `ghcr.io/ijry/pro-api`
- Prebuilt binaries in GitHub Releases
```

- [ ] **Step 6: Run a YAML parse check**

Run: `@'`nimport sys, yaml`nyaml.safe_load(open('.github/workflows/release.yml', encoding='utf-8'))`nprint('ok')`n'@ | python -`
Expected: `ok`

- [ ] **Step 7: Run the packaging helper on a fake local binary path**

Run: `New-Item -ItemType Directory -Force tmp\\release-test | Out-Null; Set-Content -Path tmp\\release-test\\proapi -Value 'stub'; bash scripts/package-release.sh linux amd64 v0.0.0 tmp/release-test/proapi tmp/release-out`
Expected: creates `tmp/release-out/proapi_linux_amd64.tar.gz` and `tmp/release-out/proapi_linux_amd64.tar.gz.sha256`

- [ ] **Step 8: Commit**

```bash
git add .github/workflows/release.yml scripts/package-release.sh README.md
git commit -m "feat(release): add release publishing pipeline"
```

### Task 2: Align installation, Docker, and upgrade documentation with real artifacts

**Files:**
- Modify: `docs-site/zh/guide/installation.md`
- Modify: `docs-site/zh/deployment/docker.md`
- Modify: `docs-site/zh/deployment/docker-compose.md`
- Modify: `docs-site/zh/deployment/reverse-proxy.md`
- Modify: `docs-site/zh/guide/upgrade.md`
- Test: `docs-site/.vitepress/dist`

**Interfaces:**
- Consumes: `.github/workflows/release.yml`, `scripts/package-release.sh`
- Produces:
  - Docs that point to `ghcr.io/ijry/pro-api`
  - Docs whose binary file names exactly match packaged release outputs
  - Install/upgrade examples that no longer describe nonexistent artifacts

- [ ] **Step 1: Write the failing doc assertions**

```text
Disallowed strings after this task:
- ghcr.io/proapi/proapi

Required strings after this task:
- ghcr.io/ijry/pro-api
- https://github.com/ijry/pro-api/releases
- proapi_linux_amd64.tar.gz
```

- [ ] **Step 2: Confirm the old registry string still exists**

Run: `rg -n "ghcr\\.io/proapi/proapi" docs-site/zh -S`
Expected: multiple matches in installation/deployment/upgrade docs

- [ ] **Step 3: Update installation page copy and commands**

```md
> pro-api 当前提供三种安装方式：Docker、预编译二进制、源码构建。

docker pull ghcr.io/ijry/pro-api:latest

docker run -d \
  --name pro-api \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -e PROAPI_MASTER_KEY=$(openssl rand -base64 32) \
  --restart unless-stopped \
  ghcr.io/ijry/pro-api:latest

wget https://github.com/ijry/pro-api/releases/download/vX.Y.Z/proapi_linux_amd64.tar.gz
sha256sum -c proapi_linux_amd64.tar.gz.sha256
```

- [ ] **Step 4: Update Docker, Compose, reverse-proxy, and upgrade pages**

```md
- Official image: `ghcr.io/ijry/pro-api:vX.Y.Z`
- Compose service image: `ghcr.io/ijry/pro-api:vX.Y.Z`
- Upgrade pull command: `docker pull ghcr.io/ijry/pro-api:vX.Y.Z`
- Rollback command: `docker run -d ... ghcr.io/ijry/pro-api:vOLD.Y.Z`
```

- [ ] **Step 5: Rebuild the docs site**

Run: `pnpm -C docs-site build`
Expected: build completes and writes `docs-site/.vitepress/dist`

- [ ] **Step 6: Search for stale registry references**

Run: `rg -n "ghcr\\.io/proapi/proapi" docs-site -S`
Expected: no matches

- [ ] **Step 7: Commit**

```bash
git add docs-site/zh/guide/installation.md docs-site/zh/deployment/docker.md docs-site/zh/deployment/docker-compose.md docs-site/zh/deployment/reverse-proxy.md docs-site/zh/guide/upgrade.md
git commit -m "docs: align install docs with published artifacts"
```

### Task 3: Unify user-visible branding from proapi to pro-api

**Files:**
- Modify: `web/admin/index.html`
- Modify: `web/user/index.html`
- Modify: `web/admin/src/layouts/AppLayout.vue`
- Modify: `web/admin/src/layouts/AuthLayout.vue`
- Modify: `web/user/src/components/biz/AppHeader.vue`
- Modify: `web/user/src/components/biz/AppFooter.vue`
- Modify: `web/shared/src/i18n/zh.json`
- Modify: `web/shared/src/i18n/en.json`
- Modify: `docs-site/index.md`
- Modify: `docs-site/zh/index.md`
- Modify: `docs-site/en/index.md`
- Modify: `docs-site/public/admin-demo/index.html`
- Modify: `docs-site/public/user-demo/index.html`
- Test: `web/admin/dist`, `web/user/dist`, `docs-site/.vitepress/dist`

**Interfaces:**
- Consumes: existing Vue/i18n/docs-site visible strings
- Produces:
  - `pro-api` in browser titles, layout headers/footers, docs hero text, and demo HTML titles
  - unchanged internal keys like `proapi.locale` and `@proapi/shared`

- [ ] **Step 1: Write the visible-branding assertions**

```text
Change:
- <title>proapi · 控制台</title> -> <title>pro-api · 控制台</title>
- brand.name: "proapi" -> "pro-api"
- visible header/footer text "proapi" -> "pro-api"

Do not change:
- proapi.locale
- proapi.theme
- proapi_csrf
- @proapi/shared
```

- [ ] **Step 2: Confirm the current visible strings are still present**

Run: `rg -n "\"proapi\"|proapi ·|>proapi<|alt: proapi|name: proapi" web docs-site -S`
Expected: matches in HTML, Vue layout/components, i18n JSON, docs-site landing pages, and demo HTML

- [ ] **Step 3: Update visible frontend and docs branding**

```html
<title>pro-api · 控制台</title>
<title>pro-api</title>
```

```vue
<NText v-if="!app.sidebarCollapsed" strong class="text-lg">pro-api</NText>
<span class="text-primary font-bold text-xl tracking-tight">pro-api</span>
<span>© 2026 pro-api · MIT</span>
```

```json
{
  "brand": {
    "name": "pro-api"
  }
}
```

```md
hero:
  name: pro-api
  image:
    alt: pro-api
```

- [ ] **Step 4: Build admin and user frontends**

Run: `pnpm -C web/admin build`
Expected: admin build passes

Run: `pnpm -C web/user build`
Expected: user build passes

- [ ] **Step 5: Rebuild docs site after branding edits**

Run: `pnpm -C docs-site build`
Expected: docs build passes

- [ ] **Step 6: Verify visible-brand replacement without touching internal keys**

Run: `rg -n "\"proapi\"|proapi ·|>proapi<|alt: proapi|name: proapi" web docs-site -S`
Expected: no matches limited to visible branding targets

Run: `rg -n "proapi\\.locale|proapi\\.theme|proapi_csrf|@proapi/" web docs-site -S`
Expected: internal identifiers still present where expected

- [ ] **Step 7: Final diff sanity and commit**

Run: `git diff --check`
Expected: no whitespace/conflict issues

```bash
git add web/admin/index.html web/user/index.html web/admin/src/layouts/AppLayout.vue web/admin/src/layouts/AuthLayout.vue web/user/src/components/biz/AppHeader.vue web/user/src/components/biz/AppFooter.vue web/shared/src/i18n/zh.json web/shared/src/i18n/en.json docs-site/index.md docs-site/zh/index.md docs-site/en/index.md docs-site/public/admin-demo/index.html docs-site/public/user-demo/index.html
git commit -m "feat(brand): unify visible pro-api branding"
```

## Self-Review

- Spec coverage:
  - Release workflow covered by Task 1.
  - Docker/binary artifact naming covered by Task 1 helper and workflow.
  - Installation/deployment/upgrade doc correction covered by Task 2.
  - User-visible `pro-api` branding only, with internal identifiers preserved, covered by Task 3.
- Placeholder scan:
  - No `TODO`/`TBD` placeholders remain.
  - Each task includes explicit files, commands, and expected outcomes.
- Type consistency:
  - Packaging helper signature is defined once in Task 1 and reused consistently.
  - GHCR registry target is consistently `ghcr.io/ijry/pro-api`.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-06-24-release-brand-installation-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?

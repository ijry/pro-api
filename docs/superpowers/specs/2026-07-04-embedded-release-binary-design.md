# Embedded Release Binary Design

## Context

The current GitHub Actions release workflow already builds platform archives and publishes a GitHub Release on `v*` tags. It also builds the admin, user, and docs frontends before running `go build -tags embed`.

The gap is runtime packaging: the Go server does not currently embed or serve the built frontend assets. The release archives contain only the executable, `LICENSE`, and `README.md`, so the binary is a backend API server rather than a complete deployable frontend + backend binary.

## Goal

Make tag-triggered Actions releases produce a single executable that serves:

- Admin frontend under `/admin`
- User frontend under `/user`
- Documentation site under `/docs`
- Existing API routes unchanged

The release artifact should remain simple to download and run, with frontend assets included in the executable produced by `go build -tags embed`.

## Non-Goals

- No frontend redesign.
- No change to API route behavior.
- No new runtime requirement for Node or pnpm.
- No hot-reload or external static asset override in this change.
- No installer or service manager packaging.

## Architecture

Add a small static asset module, for example `internal/server/static`.

The module has two build variants:

- `embed` build: uses `//go:embed` to include the built frontend directories.
- default build: exposes the same registration API but does not serve embedded frontend assets. This keeps normal local backend development from requiring frontend builds.

The command entry point calls this module after API routes are registered. The module mounts static routes on the Gin engine and leaves `/api/*`, `/v1/*`, and `/v1beta/*` behavior untouched.

## Route Behavior

The embedded binary serves three independent web roots:

- `/admin` and `/admin/*` serve `web/admin/dist`
- `/user` and `/user/*` serve `web/user/dist`
- `/docs` and `/docs/*` serve `docs-site/.vitepress/dist`

SPA fallbacks:

- Unknown paths under `/admin/*` return admin `index.html`.
- Unknown paths under `/user/*` return user `index.html`.
- Documentation paths should use the VitePress output when a file exists, then fall back to its `index.html` or the closest generated HTML page as supported by the implementation.

Root path:

- `/` redirects to `/docs/` so the binary has a useful first page without guessing whether an operator wants admin or user UI.

Caching:

- Fingerprinted assets may use long-lived cache headers.
- `index.html` responses should avoid long-lived caching so frontend deployments are not sticky after upgrades.

## Release Packaging

The workflow can keep its current build order:

1. Install frontend dependencies.
2. Build admin, user, and docs assets.
3. Build Go binary with `-tags embed`.
4. Package archives and checksums.
5. Publish GitHub Release.

`scripts/package-release.sh` should also include deployment support files that are not compiled into the binary:

- `configs/proapi.example.yaml`
- database migration files

The executable itself includes the frontend assets, so release archives do not need separate frontend dist folders.

## Error Handling

If embedded assets are missing or malformed in an `embed` build, the static module should fail fast during startup or route registration rather than silently serving blank pages.

For default non-embed builds, static frontend routes may be absent or return a clear 404. Local API development should continue to work without running frontend builds.

## Testing

Add focused tests around the static module:

- `/admin/` returns HTML in an embed build fixture.
- `/admin/some/spa/path` falls back to admin HTML.
- `/user/` and a user SPA path return HTML.
- `/docs/` returns docs HTML.
- API-like paths are not captured by frontend fallback.

Build verification:

- `pnpm -C web/admin build`
- `pnpm -C web/user build`
- `pnpm -C docs-site build`
- `go test ./...`
- `go build -tags embed -o bin/proapi ./cmd/proapi`

Release verification:

- Inspect the archive contents.
- Run the built binary with a minimal config and required dependencies.
- Verify `/`, `/admin/`, `/user/`, `/docs/`, and `/api/health` over HTTP.

## Risks

- Go embed requires files to exist at compile time. The release workflow already builds them first, but local `-tags embed` builds must do the same.
- SPA fallback must be scoped to each frontend prefix so it does not swallow API routes.
- Docs routing differs from SPA routing because VitePress emits static HTML pages; the implementation should prefer real files before falling back.

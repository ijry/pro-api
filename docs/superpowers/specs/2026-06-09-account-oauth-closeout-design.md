# Account OAuth Closeout Design

## Goal

Close the gap between the implemented account-pool OAuth PKCE flow and the project-facing documentation/configuration.

The code path is already present on `main`:

- Admin start endpoint: `POST /api/admin/accounts/oauth/start`
- No-auth callback endpoint: `GET /api/admin/accounts/oauth/callback`
- Admin UI OAuth tab opens the provider authorization popup and refreshes on completion
- OAuth providers wired for `anthropic` and `openai`

This closeout updates the operator-facing surface so the feature is discoverable, configurable, and smoke-testable.

## Scope

In scope:

- Add account OAuth and probe keys to `configs/proapi.example.yaml`.
- Update `README.md` status so account-pool OAuth onboarding is no longer listed as unfinished.
- Update `CHANGELOG.md` to describe the shipped account-pool OAuth PKCE admin flow and narrow the remaining M2 work to async tasks plus Midjourney/Suno.
- Add a short docs note for real IdP smoke testing, including required config keys and expected admin UI flow.
- Keep production behavior unchanged.

Out of scope:

- New backend config/status endpoint.
- Admin UI provider readiness badges.
- Automated real-provider E2E tests.
- Adding provider secrets; this repository only documents public client IDs/URLs and safe example values.

## Design

### Configuration

`configs/proapi.example.yaml` gets a new `account:` section matching `internal/app/config.AccountConfig`.

The section documents:

- `oauth_anthropic_auth_url`
- `oauth_anthropic_token_url`
- `oauth_anthropic_client_id`
- `oauth_anthropic_redirect_uri`
- `oauth_anthropic_scopes`
- `oauth_openai_auth_url`
- `oauth_openai_token_url`
- `oauth_openai_client_id`
- `oauth_openai_redirect_uri`
- `oauth_openai_scopes`
- `anthropic_probe_base`
- `openai_probe_base`

Values should be empty strings for deployment-specific OAuth app fields and safe example values for generic API bases.

### Documentation

README status changes:

- Keep project status as `~M2`.
- Move account-pool OAuth onboarding into ready/in-place language.
- Note that real provider onboarding requires configuring account OAuth auth/token/client/redirect values.
- Leave async task system, Midjourney, and Suno as M2 remaining work.

CHANGELOG changes:

- Add a bullet under M2 for account-pool OAuth PKCE admin onboarding.
- Update the unfinished note to remove account-pool OAuth PKCE.

Smoke note:

- Add `docs/superpowers/account-oauth-smoke.md` describing how to smoke test real IdP onboarding:
  - Set account OAuth config.
  - Ensure the provider OAuth app callback URL points to `/api/admin/accounts/oauth/callback`.
  - Start backend/admin UI.
  - Open admin account add dialog, choose OAuth tab, select channel/provider, start authorization.
  - Complete provider auth.
  - Verify a new account row appears and an `oauth_callback` event/audit entry exists.

### Error Handling Expectations

This closeout does not change runtime error behavior.

Expected failure modes remain:

- Missing provider config can generate an authorization URL with empty fields or a token exchange failure.
- Invalid/expired state returns the existing account OAuth state error.
- Provider rejection returns the existing account OAuth rejected error.

The docs should make missing config obvious before operators attempt a real smoke test.

## Testing

Because this is documentation/configuration-sample work, verification should prove no code was accidentally broken:

- `go test ./...`
- `go build ./...`
- `pnpm -C web typecheck`
- `pnpm -C web/admin build`
- `git diff --check`

If only docs/config sample files change, no real IdP smoke can be performed unless valid provider OAuth credentials are available locally.

## Acceptance Criteria

- `README.md` and `CHANGELOG.md` no longer incorrectly mark account-pool OAuth PKCE as unfinished.
- `configs/proapi.example.yaml` exposes all account OAuth/probe config keys defined in `AccountConfig`.
- A smoke-test note explains the exact required config and admin UI flow.
- Verification commands pass or any skipped real IdP smoke is explicitly documented as blocked by missing credentials.

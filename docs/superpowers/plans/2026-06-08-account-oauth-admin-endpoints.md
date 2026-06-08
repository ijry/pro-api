# Account OAuth Admin Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose the existing account-pool OAuth PKCE flow through admin HTTP endpoints and enable the admin UI to launch it.

**Architecture:** Keep PKCE/state/token exchange inside `account.OAuthFlow`. `AccountHandler` owns HTTP validation, account persistence, events, probe trigger, and audit. The callback endpoint is mounted without session auth and relies on one-time state validation.

**Tech Stack:** Go, Gin, account facade, Redis-backed OAuth state store, Vue 3, Naive UI, TypeScript.

---

### Task 1: Backend Handler Tests

**Files:**
- Modify: `internal/server/handler/admin/account_test.go`

- [ ] **Step 1: Add a fake OAuthFlow to the existing account handler test harness**

Create a fake that records `Start` input and returns a prepared account from `Callback`. Add it to the `account.Facade`.

- [ ] **Step 2: Write failing tests for start and callback**

Add tests asserting:
- `POST /api/admin/accounts/oauth/start` with `provider=openai` and `channel_id=42` returns `auth_url` and `state`.
- missing `channel_id` returns HTTP 400.
- `GET /api/admin/accounts/oauth/callback?state=s&code=c` persists the returned account, appends an `oauth_callback` event, writes `account.oauth_callback` audit, and returns HTML containing `account_oauth_done`.

- [ ] **Step 3: Verify red**

Run: `go test ./internal/server/handler/admin -run 'TestAccountHandler_OAuth'`

Expected: FAIL because `OAuthStart` and `OAuthCallback` handler methods and routes do not exist yet.

### Task 2: Backend Implementation

**Files:**
- Modify: `internal/server/handler/admin/account.go`
- Modify: `cmd/proapi/main.go`

- [ ] **Step 1: Add handler request/response code**

Implement:
- `OAuthStart`: validates JSON body, requires `channel_id > 0`, calls `h.Facade.OAuth.Start`, returns `{"data":{"auth_url": "...", "state": "..."}}`.
- `OAuthCallback`: validates `state` and `code`, calls `h.Facade.OAuth.Callback`, persists account through `Repo.Create`, appends `oauth_callback` event, optionally starts probe, audits `account.oauth_callback`, and returns a tiny HTML page that posts `account_oauth_done` to the opener and closes.

- [ ] **Step 2: Route normal admin start and no-auth callback**

In `wireAccountHandler`, mount `POST /accounts/oauth/start` inside the existing admin-authenticated account group. Mount `GET /accounts/oauth/callback` directly on `/api/admin` with only JSON error middleware already present on the parent group.

- [ ] **Step 3: Verify green**

Run: `go test ./internal/server/handler/admin -run 'TestAccountHandler_OAuth'`

Expected: PASS.

### Task 3: Frontend Minimal Wiring

**Files:**
- Modify: `web/admin/src/api/account.ts`
- Modify: `web/admin/src/views/accounts/AddDialog.vue`
- Modify: `web/admin/src/i18n/zh.json`
- Modify: `web/admin/src/i18n/en.json`

- [ ] **Step 1: Add account API methods**

Add `oauthStart(payload)` returning `{ auth_url, state }`.

- [ ] **Step 2: Enable OAuth tab**

Replace the disabled info panel with channel/provider controls and a button that calls `accountApi.oauthStart`, opens the returned URL in a popup, and listens for `account_oauth_done` to refresh and close the dialog.

- [ ] **Step 3: Verify frontend type/build health**

Run the package build command available in `web/admin/package.json`.

### Task 4: Full Verification

**Files:**
- No edits.

- [ ] **Step 1: Backend full test**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 2: Build**

Run: `go build ./...`

Expected: PASS.

- [ ] **Step 3: Manual IdP smoke note**

If real OAuth config is present, open admin, select a provider/channel, launch OAuth, complete provider auth, and verify a new account row appears. If config is absent, document the exact missing config keys.

# pro-api

All-in-one LLM API Gateway · 3-protocol interop · 18+ providers · Enterprise SSO · Full billing

[中文](./README_zh.md) · [Docs](./docs-site)

## Status

🚧 **In development · ~M2** (M0 / M1 done, most of M2 in place, M3 not started). The full local loop works: register → create token → call via any of the 3 protocols → bill → view logs → top up.

**Ready**

- **3-protocol ingress + interop**: OpenAI `/v1`, Anthropic `/v1/messages`, Gemini `/v1beta` — any ingress can route to any upstream
- **18 upstream adapters**: openai / azure / anthropic / gemini / deepseek / moonshot / zhipu / qwen / doubao / groq / mistral / yi / openrouter / huggingface / minimax / tencent / cohere / xunfei
- **Multimodal**: chat (incl. streaming) · image generation · TTS · STT · embeddings
- **Billing**: Redis Lua reserve / commit / refund · per-model ratio · per-group ratio
- **Channels**: CRUD + priority + weight + model mapping + circuit breaker; account pool
- **Rate limiting**: user / token / IP / model / group dimensions
- **Auth**: email+password + email code + session; 6 OAuth providers (GitHub / Google / WeChat / Feishu / DingTalk / Discord)
- **Payments**: Stripe / Alipay / WeChat Pay + manual top-up + redeem codes
- **Invites & commission** · **announcements** · **system settings** · **audit**
- **Frontend**: admin (full Naive UI pages) + user portal (Playground / model catalog / invites) + docs site; bilingual zh/en

**In progress / not yet done**

- Async task system (asynq) and Midjourney / Suno
- Account-pool OAuth onboarding (PKCE flow)
- M3 enterprise: SSO (OIDC / SAML / LDAP / CAS), department budgets, audit visualization, OpenTelemetry

See the roadmap at [`docs/superpowers/2026-05-21-proapi-总体路线图.md`](./docs/superpowers/2026-05-21-proapi-总体路线图.md).

## 5-minute Quickstart

For tagged releases, this repo publishes:

- Docker images at `ghcr.io/ijry/pro-api`
- Prebuilt binaries in [GitHub Releases](https://github.com/ijry/pro-api/releases)

### Prerequisites

- Go 1.22+
- Node.js 20+ and pnpm 9
- Docker (for MySQL / PostgreSQL / Redis). Podman users: see `README_zh.md` for the socket adapter snippet.

### Steps

```bash
git clone https://github.com/ijry/pro-api.git
cd pro-api
make install-tools
make docker-up
export PROAPI_MASTER_KEY=$(openssl rand -base64 32)
make dev
```

Open:

- Backend health → http://127.0.0.1:8080/healthz
- Admin → http://127.0.0.1:5173
- User → http://127.0.0.1:5174
- Docs → http://127.0.0.1:5175

## License

[MIT](./LICENSE)

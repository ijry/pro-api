# proapi

All-in-one LLM API Gateway · 3-protocol interop · 18+ providers · Enterprise SSO · Full billing

[中文](./README_zh.md) · [Docs](./docs-site)

## Status

🚧 M0 scaffolding stage — only the skeleton runs. Business features will land progressively from M1.

See the roadmap at [`docs/superpowers/2026-05-21-proapi-总体路线图.md`](./docs/superpowers/2026-05-21-proapi-总体路线图.md).

## 5-minute Quickstart

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

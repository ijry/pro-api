---
title: 反向代理
outline: deep
---

# 反向代理(Nginx / Caddy / Traefik)

> 生产环境**一定要**在 proapi 前面加反代,处理 TLS、流式 buffering、IP 透传、URI 分流。本页给出三种主流方案的完整配置示例。

## 为什么要反代

1. **TLS 终止** —— proapi 默认裸 HTTP,生产必须 HTTPS
2. **流式响应需要 disable buffering** —— 否则用户感受不到流(等所有 chunk 攒齐才收到)
3. **后台 UI / 用户 UI / API 不同路径分流** —— `/admin` `/user` `/v1` `/api` 等
4. **加 IP 限流 / WAF / CDN / fail2ban** —— 边缘防护

## 关键配置点

| 配置 | 为什么 |
|---|---|
| **SSE 流式**:关 buffering、调大 read/proxy_timeout | 否则流式响应卡住直到上游响应完 |
| **WebSocket**(M2 后用) | 实时通知需要 |
| **Header 透传**:`X-Forwarded-For` / `X-Real-IP` / `X-Request-ID` | 否则 IP 限流会全部限到反代 IP |
| **gzip**:对 JSON 响应开,SSE 关 | gzip + SSE = 灾难 |
| **client_max_body_size**:50 MB+ | Vision 请求带图,可能很大 |

## Nginx 完整配置

```nginx
upstream proapi {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;

    ssl_certificate     /etc/nginx/certs/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;

    # === 流式响应必备 ===
    proxy_buffering off;
    proxy_cache off;
    proxy_http_version 1.1;
    proxy_set_header Connection "";

    # === 上行 / 超时 ===
    client_max_body_size 50m;      # Vision 图片可能大
    proxy_read_timeout 600s;       # 流式响应可能长
    proxy_send_timeout 600s;

    # === 透传 ===
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # === 路由 ===
    # 代理 API
    location /v1/ {
        proxy_pass http://proapi;
    }

    # 管理 / 用户 / 认证 / 公开 API
    location /api/ {
        proxy_pass http://proapi;
    }

    # 健康检查
    location /healthz {
        proxy_pass http://proapi;
    }

    # Prometheus 指标:仅内网
    location /metrics {
        allow 10.0.0.0/8;
        allow 192.168.0.0/16;
        deny all;
        proxy_pass http://proapi;
    }

    # 前端静态资源(admin / user / docs)由 proapi 同实例服务
    location / {
        proxy_pass http://proapi;
    }
}

# HTTP → HTTPS 重定向
server {
    listen 80;
    server_name api.example.com;
    return 301 https://$host$request_uri;
}
```

## Caddy 完整配置

Caddy 自动 Let's Encrypt + 默认 HTTP/2,配置极简:

```caddyfile
api.example.com {
    encode gzip

    # 流式响应必备
    reverse_proxy 127.0.0.1:8080 {
        flush_interval -1
        header_up X-Forwarded-For {http.request.remote.host}
        header_up X-Real-IP {http.request.remote.host}
    }

    # /metrics 仅内网
    @internal {
        path /metrics
        not remote_ip 10.0.0.0/8 192.168.0.0/16
    }
    respond @internal 403
}
```

## Traefik(Docker labels)

适合容器化部署,直接在 compose 里加 labels:

```yaml
services:
  proapi:
    image: ghcr.io/proapi/proapi:vX.Y.Z
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.proapi.rule=Host(`api.example.com`)"
      - "traefik.http.routers.proapi.entrypoints=websecure"
      - "traefik.http.routers.proapi.tls.certresolver=le"
      - "traefik.http.services.proapi.loadbalancer.server.port=8080"

      # 流式响应:关 buffering
      - "traefik.http.middlewares.proapi-stream.buffering.maxResponseBodyBytes=0"
      - "traefik.http.routers.proapi.middlewares=proapi-stream"
```

## TLS 证书

| 方式 | 适用场景 |
|---|---|
| Let's Encrypt + certbot | Nginx,公网域名 |
| Caddy 自动 | Caddy,公网域名 |
| Traefik 自动 | Traefik,公网域名 |
| 自签 | 内网 / 开发,生产不推荐 |
| 商业证书 | 已购买的企业证书 |

## CDN

- **不建议把 `/v1/` 走 CDN** —— SSE / 计费需要 origin 精确控制,CDN 会破坏流
- **可以把 `/admin` `/user` 静态资源走 CDN** —— JS / CSS / 图片
- CDN 必须正确透传 `X-Forwarded-For` / `X-Real-IP`

## 安全加固

- `/metrics` 只对监控网段开放(白名单)
- 后台 `/api/admin/*` 加 **IP 白名单** 或 VPN-only 访问(超管账号被劫风险)
- 启用 **fail2ban** 防爆破 `/api/auth/login`(5 次失败封 IP 15 分钟)
- TLS 至少 TLSv1.2,禁 SSLv3 / TLSv1.0/1.1
- HSTS:`Strict-Transport-Security: max-age=63072000; includeSubDomains`
- `X-Content-Type-Options: nosniff` / `X-Frame-Options: DENY`

## 关键要点

- **SSE 反代最容易踩坑的是 buffering**,必须显式关
- **`proxy_read_timeout` 默认 60s**,流式响应会被截断,**调到 600+ s**
- `X-Real-IP` / `X-Forwarded-For` 必须透传,**否则 proapi 的 IP 限流会全部限到反代 IP**
- `/metrics` 不要直接公网开放(可能泄露内部指标)

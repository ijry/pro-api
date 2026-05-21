# syntax=docker/dockerfile:1.7

############################################################
# Stage 1: 前端构建
############################################################
FROM node:20-alpine AS frontend
RUN corepack enable && corepack prepare pnpm@9.12.0 --activate
WORKDIR /src
COPY web/package.json web/pnpm-workspace.yaml web/pnpm-lock.yaml web/tsconfig.base.json web/uno.config.ts web/.npmrc ./
COPY web/shared/package.json shared/
COPY web/admin/package.json admin/
COPY web/user/package.json user/
RUN pnpm install --frozen-lockfile
COPY web/shared/ shared/
COPY web/admin/ admin/
COPY web/user/ user/
RUN pnpm -C admin build && pnpm -C user build

############################################################
# Stage 2: 文档站构建
############################################################
FROM node:20-alpine AS docs
RUN corepack enable && corepack prepare pnpm@9.12.0 --activate
WORKDIR /src
COPY docs-site/package.json docs-site/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY docs-site/ ./
RUN pnpm build

############################################################
# Stage 3: Go 构建
############################################################
FROM golang:1.22-alpine AS backend
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/admin/dist /src/web/admin/dist
COPY --from=frontend /src/user/dist  /src/web/user/dist
COPY --from=docs     /src/.vitepress/dist /src/docs-site/.vitepress/dist
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 go build \
    -tags embed \
    -ldflags="-s -w \
      -X github.com/proapi/proapi/internal/version.Version=${VERSION} \
      -X github.com/proapi/proapi/internal/version.Commit=${COMMIT} \
      -X github.com/proapi/proapi/internal/version.BuildTime=${BUILD_TIME}" \
    -o /out/proapi ./cmd/proapi

############################################################
# Stage 4: 运行镜像
############################################################
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 1000 proapi
WORKDIR /app
COPY --from=backend /out/proapi /usr/local/bin/proapi
USER proapi
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/proapi"]

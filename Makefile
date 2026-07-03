# proapi 顶层 Makefile。
#
# 容器工具适配:默认 `docker compose`。podman 用户可:
#   make docker-up DOCKER_COMPOSE="podman compose"
# 或在 shell rc 中 `export DOCKER_COMPOSE="podman compose"`。

DOCKER_COMPOSE ?= docker compose

.PHONY: help dev dev-backend dev-admin dev-user dev-docs build build-backend build-frontend \
        test test-unit test-integration lint fmt clean docker-up docker-down install-tools migrate

help:
	@echo "proapi 开发命令:"
	@echo "  make install-tools     安装开发工具(golangci-lint / lefthook / pnpm 等)"
	@echo "  make docker-up         启动开发依赖(MySQL/PG/Redis)"
	@echo "  make docker-down       停止依赖"
	@echo "  make migrate           跑数据库迁移"
	@echo "  make dev               并发起后端 + admin + user + docs(需先 docker-up)"
	@echo "  make dev-backend       仅起 Go 后端"
	@echo "  make dev-admin         仅起 admin 前端"
	@echo "  make dev-user          仅起 user 前端"
	@echo "  make dev-docs          仅起 docs-site"
	@echo "  make build             构建生产产物(嵌入式单二进制)"
	@echo "  make test              全量测试"
	@echo "  make lint              所有 lint"
	@echo "  make fmt               所有格式化"
	@echo "  make clean             清理产物"
	@echo ""
	@echo "容器工具变量 DOCKER_COMPOSE=\"$(DOCKER_COMPOSE)\"(podman 用户可改 \"podman compose\")"

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
	go install mvdan.cc/gofumpt@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@command -v pnpm >/dev/null 2>&1 || npm install -g pnpm@9
	pnpm dlx lefthook install

docker-up:
	$(DOCKER_COMPOSE) -f deploy/docker-compose.dev.yml up -d

docker-down:
	$(DOCKER_COMPOSE) -f deploy/docker-compose.dev.yml down

migrate:
	./scripts/migrate.sh up

dev-backend:
	go run ./cmd/proapi

dev-admin:
	pnpm -C web/admin dev

dev-user:
	pnpm -C web/user dev

dev-docs:
	pnpm -C docs-site dev

dev:
	@command -v concurrently >/dev/null 2>&1 || (echo "需要 concurrently: pnpm i -g concurrently" && exit 1)
	concurrently -n backend,admin,user,docs -c blue,green,magenta,yellow \
		"make dev-backend" "make dev-admin" "make dev-user" "make dev-docs"

build-frontend:
	PROAPI_ADMIN_BASE=/admin/ pnpm -C web/admin build
	PROAPI_USER_BASE=/user/ pnpm -C web/user build
	DOCS_BASE=/docs/ pnpm -C docs-site build

build-backend:
	bash scripts/prepare-embed-assets.sh
	CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o bin/proapi ./cmd/proapi

build: build-frontend build-backend

test-unit:
	go test -race -count=1 ./...

test-integration:
	go test -race -count=1 -tags=integration ./...

test: test-unit test-integration

lint:
	golangci-lint run ./...
	pnpm -C web lint
	pnpm -C web typecheck

fmt:
	gofumpt -l -w .
	pnpm -C web format

clean:
	rm -rf bin/ dist/
	rm -rf web/admin/dist web/user/dist docs-site/.vitepress/dist
	rm -rf web/admin/.vite web/user/.vite docs-site/.vitepress/cache

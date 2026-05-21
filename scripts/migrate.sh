#!/usr/bin/env bash
# proapi 数据库迁移包装脚本。
#
# 用法:
#   ./scripts/migrate.sh up
#   ./scripts/migrate.sh down 1
#   ./scripts/migrate.sh version
#
# 读取环境变量:
#   PROAPI_DATABASE_DRIVER   mysql | postgres
#   PROAPI_DATABASE_DSN      数据库连接串
#
# 需要预装 golang-migrate:
#   go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
set -euo pipefail

DRIVER="${PROAPI_DATABASE_DRIVER:-mysql}"
DSN="${PROAPI_DATABASE_DSN:-}"

if [[ -z "$DSN" ]]; then
  echo "ERROR: PROAPI_DATABASE_DSN is required" >&2
  exit 1
fi

case "$DRIVER" in
  mysql)
    SOURCE="file://migrations/mysql"
    MIGRATE_DSN="mysql://${DSN}"
    ;;
  postgres)
    SOURCE="file://migrations/postgres"
    MIGRATE_DSN="${DSN}"
    ;;
  *)
    echo "ERROR: unsupported driver: $DRIVER" >&2
    exit 1
    ;;
esac

exec migrate -source "$SOURCE" -database "$MIGRATE_DSN" "$@"

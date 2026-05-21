# 数据库迁移

工具:[golang-migrate](https://github.com/golang-migrate/migrate)。
两套并行维护:`mysql/` 与 `postgres/`,文件名前缀相同。

## 命名

`NNNNNN_<动作>_<对象>.{up,down}.sql`,序号 6 位,严格递增,不允许改历史。

## 命令

通过 `scripts/migrate.sh` 包装,自动读 `PROAPI_DATABASE_DRIVER` 与 `PROAPI_DATABASE_DSN`:

```bash
./scripts/migrate.sh up
./scripts/migrate.sh down 1
./scripts/migrate.sh version
```

## 双库注意事项

- 时间类型:MySQL `DATETIME` ↔ PG `TIMESTAMPTZ`,业务代码统一用 `time.Time` UTC
- 布尔:MySQL `TINYINT(1)` ↔ PG `BOOLEAN`
- JSON:MySQL `JSON` ↔ PG `JSONB`,GORM tag 写 `type:json`
- 自增/标识列:proapi 业务表用 Snowflake,migration 内不要 `AUTO_INCREMENT` / `IDENTITY`

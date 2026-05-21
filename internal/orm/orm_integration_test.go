//go:build integration

package orm

import (
	"fmt"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/proapi/proapi/internal/app/config"
)

func mustPool(t *testing.T) *dockertest.Pool {
	t.Helper()
	p, err := dockertest.NewPool("")
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	if err := p.Client.Ping(); err != nil {
		t.Skipf("docker daemon not reachable: %v", err)
	}
	return p
}

func TestOpen_MySQL(t *testing.T) {
	pool := mustPool(t)
	res, err := pool.Run("mysql", "8.0", []string{
		"MYSQL_ROOT_PASSWORD=proapi",
		"MYSQL_DATABASE=proapi",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Purge(res) })

	dsn := fmt.Sprintf("root:proapi@tcp(127.0.0.1:%s)/proapi?charset=utf8mb4&parseTime=True&loc=UTC",
		res.GetPort("3306/tcp"))

	pool.MaxWait = 90 * time.Second
	if err := pool.Retry(func() error {
		db, err := Open(config.DatabaseConfig{
			Driver: "mysql", DSN: dsn,
			MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifeSec: 60,
		})
		if err != nil {
			return err
		}
		sqlDB, _ := db.DB()
		return sqlDB.Ping()
	}); err != nil {
		t.Fatal(err)
	}
}

func TestOpen_Postgres(t *testing.T) {
	pool := mustPool(t)
	res, err := pool.Run("postgres", "16", []string{
		"POSTGRES_PASSWORD=proapi",
		"POSTGRES_DB=proapi",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Purge(res) })

	dsn := fmt.Sprintf("postgres://postgres:proapi@127.0.0.1:%s/proapi?sslmode=disable",
		res.GetPort("5432/tcp"))

	pool.MaxWait = 60 * time.Second
	if err := pool.Retry(func() error {
		db, err := Open(config.DatabaseConfig{
			Driver: "postgres", DSN: dsn,
			MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifeSec: 60,
		})
		if err != nil {
			return err
		}
		sqlDB, _ := db.DB()
		return sqlDB.Ping()
	}); err != nil {
		t.Fatal(err)
	}
}


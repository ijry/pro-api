package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "proapi.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PROAPI_MASTER_KEY", "0123456789abcdef0123456789abcdef")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Addr != ":8080" {
		t.Errorf("want default addr :8080, got %q", cfg.Server.Addr)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("want default log level info, got %q", cfg.Log.Level)
	}
}

func TestLoad_FromFile(t *testing.T) {
	t.Setenv("PROAPI_MASTER_KEY", "0123456789abcdef0123456789abcdef")
	p := writeTempYAML(t, `
server:
  addr: :9090
log:
  level: debug
database:
  driver: postgres
  dsn: postgres://localhost/p
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Addr != ":9090" {
		t.Errorf("got %q", cfg.Server.Addr)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("got %q", cfg.Log.Level)
	}
	if cfg.Database.Driver != "postgres" {
		t.Errorf("got %q", cfg.Database.Driver)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	t.Setenv("PROAPI_MASTER_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("PROAPI_SERVER_ADDR", ":7777")
	p := writeTempYAML(t, "server:\n  addr: :9090\n")
	cfg, _ := Load(p)
	if cfg.Server.Addr != ":7777" {
		t.Errorf("env should override file: got %q", cfg.Server.Addr)
	}
}

func TestLoad_MissingMasterKey(t *testing.T) {
	t.Setenv("PROAPI_MASTER_KEY", "")
	if _, err := Load(""); err == nil {
		t.Fatal("want error when PROAPI_MASTER_KEY is empty")
	}
}

func TestLoad_RejectsInvalidDriver(t *testing.T) {
	t.Setenv("PROAPI_MASTER_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("PROAPI_DATABASE_DRIVER", "oracle")
	if _, err := Load(""); err == nil {
		t.Fatal("want error for unsupported driver")
	}
}

package orm

import (
	"testing"

	"github.com/proapi/proapi/internal/app/config"
)

func TestOpen_RejectsUnsupportedDriver(t *testing.T) {
	if _, err := Open(config.DatabaseConfig{Driver: "sqlite"}); err == nil {
		t.Fatal("want error for sqlite")
	}
}

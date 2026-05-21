package app

import (
	"context"
	"errors"
	"testing"

	"github.com/ijry/pro-api/internal/app/config"
	"go.uber.org/zap"
)

func TestSetupBasic_FailsOnBadMasterKey(t *testing.T) {
	cfg := &config.Config{
		MasterKey: "too-short",
		NodeID:    0,
	}
	_, err := SetupBasic(context.Background(), cfg, zap.NewNop())
	if err == nil {
		t.Fatal("want error for short master_key")
	}
}

func TestSetupBasic_FailsOnBadNodeID(t *testing.T) {
	cfg := &config.Config{
		MasterKey: "0123456789abcdef0123456789abcdef",
		NodeID:    9999,
	}
	_, err := SetupBasic(context.Background(), cfg, zap.NewNop())
	if err == nil {
		t.Fatal("want error for invalid node_id")
	}
}

func TestAddCloser_LIFOOrder(t *testing.T) {
	a := &Application{}
	var order []string
	a.AddCloser("first", func() error {
		order = append(order, "first")
		return nil
	})
	a.AddCloser("second", func() error {
		order = append(order, "second")
		return nil
	})
	a.AddCloser("third", func() error {
		order = append(order, "third")
		return nil
	})
	_ = a.Shutdown(context.Background())
	want := []string{"third", "second", "first"}
	for i, w := range want {
		if order[i] != w {
			t.Fatalf("order[%d] = %s, want %s", i, order[i], w)
		}
	}
}

func TestShutdown_ContinuesOnCloserError(t *testing.T) {
	a := &Application{}
	called := []string{}
	a.AddCloser("a", func() error {
		called = append(called, "a")
		return errors.New("boom")
	})
	a.AddCloser("b", func() error {
		called = append(called, "b")
		return nil
	})
	err := a.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected non-nil error from at least one closer")
	}
	if len(called) != 2 {
		t.Fatalf("expected 2 closers called, got %d (%v)", len(called), called)
	}
}

func TestDecodeMasterKey_Hex(t *testing.T) {
	k, err := decodeMasterKey("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != 32 {
		t.Fatalf("want 32 bytes, got %d", len(k))
	}
}

func TestDecodeMasterKey_Base64(t *testing.T) {
	k, err := decodeMasterKey("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != 32 {
		t.Fatalf("want 32 bytes, got %d", len(k))
	}
}

func TestDecodeMasterKey_Raw(t *testing.T) {
	k, err := decodeMasterKey("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	if len(k) != 32 {
		t.Fatalf("want 32 bytes, got %d", len(k))
	}
}

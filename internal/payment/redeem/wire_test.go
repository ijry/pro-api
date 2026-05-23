package redeem

import (
	"context"
	"testing"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/payment"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type wireFakeWallet struct{}

func (wireFakeWallet) Credit(ctx context.Context, userID, amount int64, refType string, refID int64, desc string) error {
	return nil
}

type wireFakeSetting struct{}

func (wireFakeSetting) GetInt(ctx context.Context, k string, d int) int { return d }

func newWireApp(t *testing.T) *app.Application {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	idg, err := idgen.New(2)
	if err != nil {
		t.Fatalf("idgen: %v", err)
	}
	return &app.Application{
		DB:    db,
		IDGen: idg,
		Audit: audit.NewNoop(),
		Log:   zap.NewNop(),
		Clock: clock.Real,
	}
}

func TestWire_Redeem_NilApp_ReturnsError(t *testing.T) {
	if err := Wire(nil); err == nil {
		t.Fatal("want error")
	}
}

func TestWire_Redeem_MissingDB_ReturnsError(t *testing.T) {
	a := &app.Application{}
	if err := Wire(a); err == nil {
		t.Fatal("want error")
	}
}

func TestWire_Redeem_Success_PopulatesHolder(t *testing.T) {
	a := newWireApp(t)
	svc := New(Config{
		DB:      a.DB,
		Setting: wireFakeSetting{},
		Wallet:  wireFakeWallet{},
		Audit:   a.Audit,
		IDGen:   a.IDGen,
		Clock:   a.Clock,
		Log:     a.Log,
	})
	h := payment.HolderFrom(a.PaymentSvc)
	h.Redeem = svc
	a.PaymentSvc = h

	got := ServiceFrom(a)
	if got != svc {
		t.Fatalf("ServiceFrom mismatch: %T vs %T", got, svc)
	}
}

func TestServiceFrom_NotWired_ReturnsNil(t *testing.T) {
	a := &app.Application{}
	if ServiceFrom(a) != nil {
		t.Fatal("want nil")
	}
	if ServiceFrom(nil) != nil {
		t.Fatal("want nil for nil app")
	}
}

func TestHolder_ManualAndRedeem_Coexist(t *testing.T) {
	// 模拟 manual 已经被装配
	pre := &payment.Holder{Manual: "manual-svc"}
	a := newWireApp(t)
	a.PaymentSvc = pre

	svc := New(Config{
		DB:      a.DB,
		Setting: wireFakeSetting{},
		Wallet:  wireFakeWallet{},
		Audit:   a.Audit,
		IDGen:   a.IDGen,
		Clock:   a.Clock,
		Log:     a.Log,
	})
	h := payment.HolderFrom(a.PaymentSvc)
	h.Redeem = svc
	a.PaymentSvc = h

	// 复用前一个 holder,manual 字段保留
	h2, ok := a.PaymentSvc.(*payment.Holder)
	if !ok || h2 == nil {
		t.Fatal("holder lost")
	}
	if h2.Manual != "manual-svc" {
		t.Fatalf("Manual lost: %v", h2.Manual)
	}
	if h2.Redeem == nil {
		t.Fatal("Redeem missing")
	}
}

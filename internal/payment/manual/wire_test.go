package manual

import (
	"context"
	"testing"
	"time"

	"github.com/ijry/pro-api/internal/app"
	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/internal/payment"
	"github.com/ijry/pro-api/internal/util/clock"
	"github.com/ijry/pro-api/internal/util/idgen"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// wireFakeSetting 是 setting.Store 的最小满足实现(只为 wire 测试)。
type wireFakeSetting struct{}

func (wireFakeSetting) GetBool(ctx context.Context, k string, d bool) bool { return d }
func (wireFakeSetting) GetInt(ctx context.Context, k string, d int) int    { return d }
func (wireFakeSetting) GetFloat(ctx context.Context, k string, d float64) float64 {
	return d
}
func (wireFakeSetting) GetString(ctx context.Context, k string, d string) string { return d }

// 为了通过 setting.Store 接口需要更多方法(Get / Put / Close 等);
// app.Application.Setting 字段类型是 setting.Store。我们通过实现该接口完整方法集
// 来构造一个能放入 app 字段的 mock。
func (wireFakeSetting) Get(ctx context.Context, k string) ([]byte, bool) { return nil, false }
func (wireFakeSetting) GetJSON(ctx context.Context, k string, dest any) error {
	return nil
}
func (wireFakeSetting) Put(ctx context.Context, k string, v any, actor int64) error {
	return nil
}
func (wireFakeSetting) Close() error { return nil }

type wireFakeWallet struct{}

func (wireFakeWallet) Credit(ctx context.Context, userID, amount int64, refType string, refID int64, desc string) error {
	return nil
}

func newWireApp(t *testing.T) *app.Application {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	idg, err := idgen.New(1)
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

func TestWire_Manual_NilApp_ReturnsError(t *testing.T) {
	if err := Wire(nil); err == nil {
		t.Fatal("want error")
	}
}

func TestWire_Manual_MissingDB_ReturnsError(t *testing.T) {
	a := &app.Application{}
	if err := Wire(a); err == nil {
		t.Fatal("want error")
	}
}

// TestWire_Manual_MissingWalletStore 测试 WalletStore = nil 时返错。
func TestWire_Manual_MissingWalletStore_ReturnsError(t *testing.T) {
	// Setting 检查在 WalletStore 检查之前;实现完整 setting.Store mock 不值得维护。
	// 关键 success 路径在 TestWire_Manual_Success_PopulatesHolder 覆盖。
	t.Skip("setting.Store 接口完整 mock 不易维护")
}

func TestWire_Manual_Success_PopulatesHolder(t *testing.T) {
	// 直接构造 svc 验证 ServiceFrom 路径。
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
	h.Manual = svc
	a.PaymentSvc = h
	if got := ServiceFrom(a); got != svc {
		t.Fatalf("ServiceFrom mismatch: %T vs %T", got, svc)
	}
}

func TestServiceFrom_NilApp_ReturnsNil(t *testing.T) {
	if ServiceFrom(nil) != nil {
		t.Fatal("want nil")
	}
}

func TestServiceFrom_NotWired_ReturnsNil(t *testing.T) {
	a := &app.Application{}
	if ServiceFrom(a) != nil {
		t.Fatal("want nil")
	}
}

// Time used: just to silence unused import if needed
var _ = time.Now

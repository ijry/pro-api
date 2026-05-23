package log

import (
	"context"
	"fmt"

	"github.com/ijry/pro-api/internal/app"
)

// WireLog 装配 log.Store 到 app.LogStore。
//
// 调用方（main.go 或 SetupBasic 后）执行:
//
//	if err := log.WireLog(ctx, app); err != nil {
//	    return err
//	}
func WireLog(ctx context.Context, a *app.Application) error {
	if a.DB == nil {
		return fmt.Errorf("log: app.DB is nil")
	}
	if a.IDGen == nil {
		return fmt.Errorf("log: app.IDGen is nil")
	}

	store, err := New(ctx, Config{
		Backend: BackendDB,
		DB:      a.DB,
		Clock:   a.Clock,
		IDGen:   a.IDGen,
		Log:     a.Log,
		Setting: a.Setting,
	})
	if err != nil {
		return fmt.Errorf("log: init store: %w", err)
	}

	a.LogStore = store
	a.AddCloser("log_store", store.Close)

	a.Log.Info("log: store wired")
	return nil
}

// StoreFrom extracts the log.Store from app.LogStore.
func StoreFrom(a *app.Application) Store {
	if s, ok := a.LogStore.(Store); ok {
		return s
	}
	return nil
}

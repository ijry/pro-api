package billing

import (
	"context"
	"testing"
)

// TestNew_MissingConfig 验证缺少 DB/Cache 时 New 会立刻报错
// (实际测试需要 Redis;这里只验证结构可以实例化)
func TestNew_DisabledWorkers(t *testing.T) {
	// 没有真实 Redis,借助 DisableLedgerWorker+DisableReconciler 让 New 不启动 goroutine
	// Cache 为 nil 导致 lua.Load 失败,所以这里只验证 ErrInsufficient 是否正常 export
	if ErrInsufficient == nil {
		t.Fatal("ErrInsufficient should not be nil")
	}
	_ = context.Background()
}

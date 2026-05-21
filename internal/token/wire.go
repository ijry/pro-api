package token

import (
	"github.com/ijry/pro-api/internal/app"
)

// WireToken 把 token.Store 装到 app.Application。
//
// 调用顺序:必须在 app.SetupBasic 之后(需要 DB/Cache/IDGen/Audit 都已就绪)。
//
// 注入到 app.TokenStore(any)— 其他模块通过类型断言取出 Store 接口,
// 或定义自己的最小接口(如 billing.UsageIncrementer)做反向依赖解耦(spec §11.11)。
//
// 关停在 app.Shutdown 通过 closer 链调用 Store.Close()(final flush + Pub/Sub 退订)。
func WireToken(application *app.Application) error {
	store, err := New(Config{
		DB:            application.DB,
		Cache:         application.Cache,
		Log:           application.Log,
		Clock:         application.Clock,
		IDGen:         application.IDGen,
		Audit:         application.Audit,
		// 其余字段使用默认(spec §8 setting 项目前不读取,M2 再接 setting.Store)
	})
	if err != nil {
		return err
	}
	application.TokenStore = store
	application.AddCloser("token", store.Close)
	if application.Log != nil {
		application.Log.Info("token store wired")
	}
	return nil
}

package ratelimit

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/ijry/pro-api/internal/app"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// settingInvalidateChannel 与 internal/setting 内部常量一致;复用同一 Pub/Sub channel。
const settingInvalidateChannel = "proapi:setting:invalidate"

// WireRateLimit 装配 limiter + planner 到 app.Application,并启动 setting 失效订阅。
//
// 顺序要求:在 SetupBasic 之后(需要 Cache / Setting / Log),server 路由挂载之前。
func WireRateLimit(a *app.Application) error {
	if a == nil {
		return errors.New("ratelimit: application is nil")
	}
	if a.Cache == nil {
		return errors.New("ratelimit: app.Cache is nil")
	}
	if a.Setting == nil {
		return errors.New("ratelimit: app.Setting is nil")
	}
	log := a.Log
	if log == nil {
		log = zap.NewNop()
	}
	l, err := newRedisLimiter(context.Background(), Config{
		Cache: a.Cache,
		Log:   log,
		Clock: a.Clock,
	})
	if err != nil {
		return err
	}
	planner := NewPlanner(PlannerConfig{Setting: a.Setting})

	w := &subWatcher{
		rdb:     a.Cache,
		log:     log,
		planner: planner,
		stopCh:  make(chan struct{}),
	}
	w.start()

	a.Limiter = l
	// Planner 没有 Application 占位字段,挂到 LogStore 类型?不行 — 由于 application.go 不可改,
	// 这里通过 closures 在 Middleware 调用侧自行透传。为方便集成方使用,我们暴露 PlannerFrom helper。
	plannerRegistryMu.Lock()
	plannerRegistry[a] = planner
	plannerRegistryMu.Unlock()

	a.AddCloser("ratelimit", func() error {
		w.stop()
		plannerRegistryMu.Lock()
		delete(plannerRegistry, a)
		plannerRegistryMu.Unlock()
		return l.Close()
	})
	log.Info("ratelimit wired (limiter + planner + setting watcher)")
	return nil
}

// PlannerFrom 取 application 关联的 Planner。
// 集成方在挂中间件时调用:ratelimit.Middleware(l, ratelimit.PlannerFrom(app), app.Setting, app.Log)
func PlannerFrom(a *app.Application) *Planner {
	if a == nil {
		return nil
	}
	plannerRegistryMu.RLock()
	defer plannerRegistryMu.RUnlock()
	return plannerRegistry[a]
}

var (
	plannerRegistry   = map[*app.Application]*Planner{}
	plannerRegistryMu sync.RWMutex
)

// subWatcher 订阅 setting Pub/Sub channel,收到 ratelimit.* key 时清空 planner 缓存。
type subWatcher struct {
	rdb     *redis.Client
	log     *zap.Logger
	planner *Planner
	sub     *redis.PubSub
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func (w *subWatcher) start() {
	if w.rdb == nil {
		return
	}
	w.sub = w.rdb.Subscribe(context.Background(), settingInvalidateChannel)
	w.wg.Add(1)
	go w.loop()
}

func (w *subWatcher) loop() {
	defer w.wg.Done()
	ch := w.sub.Channel()
	for {
		select {
		case <-w.stopCh:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if strings.HasPrefix(msg.Payload, "ratelimit.") {
				w.planner.InvalidateCache()
				w.log.Debug("ratelimit: planner cache purged by setting change",
					zap.String("key", msg.Payload))
			}
		}
	}
}

func (w *subWatcher) stop() {
	select {
	case <-w.stopCh:
	default:
		close(w.stopCh)
	}
	if w.sub != nil {
		_ = w.sub.Close()
	}
	w.wg.Wait()
}

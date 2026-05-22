package adapter

import (
	"fmt"
	"sort"
	"sync"
)

// Registry 按 provider 名(string)管理 Adapter 实例。
type Registry interface {
	Register(a Adapter)
	Get(name string) (Adapter, bool)
	MustGet(name string) Adapter
	List() []Adapter
	Names() []string
}

type registry struct {
	mu sync.RWMutex
	m  map[string]Adapter
}

// NewRegistry 构造空注册表。
func NewRegistry() Registry { return &registry{m: map[string]Adapter{}} }

// Register 注册一个 adapter。同名重复注册会 panic(M1 仅启动期注册)。
func (r *registry) Register(a Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.m[a.Name()]; exists {
		panic(fmt.Sprintf("adapter %s already registered", a.Name()))
	}
	r.m[a.Name()] = a
}

// Get 取一个 adapter。
func (r *registry) Get(name string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.m[name]
	return a, ok
}

// MustGet 取一个 adapter,不存在则 panic(仅 wire 阶段配置错误时触发)。
func (r *registry) MustGet(name string) Adapter {
	a, ok := r.Get(name)
	if !ok {
		panic("unknown adapter: " + name)
	}
	return a
}

// List 返回所有已注册 adapter,按 name 升序。
func (r *registry) List() []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Adapter, 0, len(r.m))
	for _, a := range r.m {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// Names 返回所有 adapter 名,按字典序升序。
func (r *registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.m))
	for n := range r.m {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

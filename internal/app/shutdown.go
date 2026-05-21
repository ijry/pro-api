package app

import "sync"

// closer 是一个关停回调。
type closer struct {
	name string
	fn   func() error
}

// closers 在 Application 里嵌入,管理 LIFO 关停列表。
type closers struct {
	mu    sync.Mutex
	items []closer
}

func (c *closers) Add(name string, fn func() error) {
	c.mu.Lock()
	c.items = append(c.items, closer{name: name, fn: fn})
	c.mu.Unlock()
}

// Run 按 LIFO 调用所有 closer;一个失败不影响其他;返回最后一个非 nil error。
func (c *closers) Run() error {
	c.mu.Lock()
	items := c.items
	c.items = nil
	c.mu.Unlock()
	var last error
	for i := len(items) - 1; i >= 0; i-- {
		if err := items[i].fn(); err != nil {
			last = err
		}
	}
	return last
}

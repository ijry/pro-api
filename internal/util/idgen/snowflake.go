// Package idgen 提供分布式 ID 生成。当前实现:Snowflake(64-bit)。
//
// 位分布:1 sign + 41 ms timestamp + 10 worker_id + 12 sequence
package idgen

import (
	"fmt"
	"sync"
	"time"
)

const (
	epoch       int64 = 1704067200000 // 2024-01-01 00:00:00 UTC,ms
	workerBits        = 10
	seqBits           = 12
	workerShift       = seqBits
	timeShift         = seqBits + workerBits
	maxWorker         = (1 << workerBits) - 1
	maxSeq      int64 = (1 << seqBits) - 1
)

// Generator 是线程安全的 Snowflake 实现。
type Generator struct {
	mu       sync.Mutex
	workerID int64
	lastMs   int64
	seq      int64
}

// New 构造一个 Generator;workerID 必须在 [0, 1023]。
func New(workerID int) (*Generator, error) {
	if workerID < 0 || workerID > maxWorker {
		return nil, fmt.Errorf("idgen: workerID out of range [0, %d]: %d", maxWorker, workerID)
	}
	return &Generator{workerID: int64(workerID)}, nil
}

// Generate 返回一个全局唯一、单调递增的 64-bit ID。
func (g *Generator) Generate() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now().UnixMilli()
	if now == g.lastMs {
		g.seq = (g.seq + 1) & maxSeq
		if g.seq == 0 {
			// 同毫秒内 seq 用尽,自旋到下一毫秒
			for now <= g.lastMs {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.seq = 0
	}
	g.lastMs = now
	return ((now - epoch) << timeShift) | (g.workerID << workerShift) | g.seq
}

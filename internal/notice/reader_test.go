package notice

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newReader(t *testing.T) (*reader, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return &reader{rdb: rdb}, mr
}

func TestReader_MarkRead_IsIdempotent(t *testing.T) {
	r, _ := newReader(t)
	ctx := context.Background()
	if err := r.MarkRead(ctx, 1, 100); err != nil {
		t.Fatal(err)
	}
	// 重复
	if err := r.MarkRead(ctx, 1, 100); err != nil {
		t.Fatal(err)
	}
	ok, err := r.IsRead(ctx, 1, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("want IsRead true")
	}
}

func TestReader_IsRead_FalseWhenNotMarked(t *testing.T) {
	r, _ := newReader(t)
	ok, err := r.IsRead(context.Background(), 1, 999)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("want false")
	}
}

func TestReader_ReadSet_ReturnsAllMembers(t *testing.T) {
	r, _ := newReader(t)
	ctx := context.Background()
	_ = r.MarkRead(ctx, 1, 100)
	_ = r.MarkRead(ctx, 1, 200)
	_ = r.MarkRead(ctx, 1, 300)
	set, err := r.ReadSet(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(set) != 3 {
		t.Fatalf("want 3, got %d", len(set))
	}
	for _, id := range []int64{100, 200, 300} {
		if _, ok := set[id]; !ok {
			t.Fatalf("missing %d", id)
		}
	}
}

func TestReader_UnreadCount_AllUnread(t *testing.T) {
	r, _ := newReader(t)
	n, err := r.UnreadCount(context.Background(), 1, []int64{10, 20, 30})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("want 3, got %d", n)
	}
}

func TestReader_UnreadCount_PartiallyRead(t *testing.T) {
	r, _ := newReader(t)
	ctx := context.Background()
	_ = r.MarkRead(ctx, 1, 20)
	n, err := r.UnreadCount(ctx, 1, []int64{10, 20, 30})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("want 2 unread, got %d", n)
	}
}

func TestReader_UnreadCount_LargeSet_Batched(t *testing.T) {
	r, _ := newReader(t)
	ctx := context.Background()
	// 准备 1200 个候选 id
	ids := make([]int64, 1200)
	for i := range ids {
		ids[i] = int64(1000 + i)
	}
	// 标 600 为已读
	for i := 0; i < 600; i++ {
		_ = r.MarkRead(ctx, 1, ids[i])
	}
	n, err := r.UnreadCount(ctx, 1, ids)
	if err != nil {
		t.Fatal(err)
	}
	if n != 600 {
		t.Fatalf("want 600 unread (1200-600), got %d", n)
	}
}

func TestReader_UnreadCount_EmptyVisible_ReturnsZero(t *testing.T) {
	r, _ := newReader(t)
	n, err := r.UnreadCount(context.Background(), 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("want 0, got %d", n)
	}
}

func TestReader_ReadKeyFormat(t *testing.T) {
	if k := readKey(42); k != "notice:read:42" {
		t.Fatalf("got %q", k)
	}
}

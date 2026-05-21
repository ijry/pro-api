package idgen

import (
	"sync"
	"testing"
)

func TestNew_AcceptsValidNodeID(t *testing.T) {
	if _, err := New(0); err != nil {
		t.Fatal(err)
	}
	if _, err := New(1023); err != nil {
		t.Fatal(err)
	}
}

func TestNew_RejectsInvalidNodeID(t *testing.T) {
	if _, err := New(-1); err == nil {
		t.Fatal("want error for -1")
	}
	if _, err := New(1024); err == nil {
		t.Fatal("want error for 1024")
	}
}

func TestGenerate_Monotonic(t *testing.T) {
	g, _ := New(1)
	prev := g.Generate()
	for i := 0; i < 10000; i++ {
		id := g.Generate()
		if id <= prev {
			t.Fatalf("not monotonic at i=%d: %d <= %d", i, id, prev)
		}
		prev = id
	}
}

func TestGenerate_Unique_Concurrent(t *testing.T) {
	g, _ := New(2)
	const N = 10000
	out := make(chan int64, N)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < N/10; j++ {
				out <- g.Generate()
			}
		}()
	}
	wg.Wait()
	close(out)
	seen := make(map[int64]struct{}, N)
	for id := range out {
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id %d", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != N {
		t.Fatalf("expect %d uniq ids, got %d", N, len(seen))
	}
}

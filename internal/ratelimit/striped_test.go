package ratelimit

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestStripedMatchesUnstriped is model-based: the striped limiter and the
// original must make the same allow/deny decisions for the same key sequence.
// Striping is a locking change, not a behaviour change.
func TestStripedMatchesUnstriped(t *testing.T) {
	const rate, burst = 10, 5
	period := time.Second

	orig := NewTokenBucket(rate, burst, period)
	strp := NewStripedTokenBucket(rate, burst, period, 64)

	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("user%d", i%7)
		want := orig.Allow(key)
		got := strp.Allow(key)
		if got != want {
			t.Fatalf("op %d key %s: striped=%v, unstriped=%v", i, key, got, want)
		}
	}
}

// TestBurstThenDeny pins the actual limiting behaviour rather than trusting it.
func TestBurstThenDeny(t *testing.T) {
	s := NewStripedTokenBucket(1, 3, time.Hour, 16) // effectively no refill
	allowed := 0
	for i := 0; i < 10; i++ {
		if s.Allow("same-key") {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("allowed %d of 10, want exactly burst=3", allowed)
	}
}

// TestKeysAreIndependent guards the property that makes striping safe: no
// cross-key invariants. Exhausting one key must not affect another.
func TestKeysAreIndependent(t *testing.T) {
	s := NewStripedTokenBucket(1, 2, time.Hour, 32)
	for i := 0; i < 5; i++ {
		s.Allow("noisy")
	}
	if !s.Allow("quiet") {
		t.Fatal("exhausting one key must not deny a different key")
	}
}

// TestConcurrentAllow is the reason the type exists. Run with -race.
func TestConcurrentAllow(t *testing.T) {
	s := NewStripedTokenBucket(1000, 1000, time.Second, 256)
	const goroutines, iters = 32, 500

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				s.Allow(fmt.Sprintf("ip-%d", (g*iters+i)%2000))
				s.Remaining(fmt.Sprintf("ip-%d", i%2000))
			}
		}(g)
	}
	wg.Wait()

	if n := s.Len(); n == 0 {
		t.Fatal("expected tracked keys after concurrent load")
	}
}

// TestReapBoundsMemory pins the leak fix: without reaping, per-IP limiting
// grows one entry per distinct address forever.
func TestReapBoundsMemory(t *testing.T) {
	s := NewStripedTokenBucket(10, 10, time.Second, 16)
	for i := 0; i < 500; i++ {
		s.Allow(fmt.Sprintf("ip-%d", i))
	}
	if got := s.Len(); got != 500 {
		t.Fatalf("Len=%d, want 500", got)
	}
	time.Sleep(20 * time.Millisecond)
	removed := s.Reap(10 * time.Millisecond)
	if removed != 500 || s.Len() != 0 {
		t.Fatalf("reaped %d leaving %d, want 500 leaving 0", removed, s.Len())
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: go test -bench=Allow -benchtime=200000x -cpu=1,2,4,8 ./internal/ratelimit/
// ---------------------------------------------------------------------------

func benchAllow(b *testing.B, limiter Limiter) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			limiter.Allow(fmt.Sprintf("ip-%d", i%2000))
			i++
		}
	})
}

func BenchmarkAllowSingleLock(b *testing.B) {
	benchAllow(b, NewTokenBucket(1e9, 1e9, time.Second))
}
func BenchmarkAllowStriped8(b *testing.B) {
	benchAllow(b, NewStripedTokenBucket(1e9, 1e9, time.Second, 8))
}
func BenchmarkAllowStriped64(b *testing.B) {
	benchAllow(b, NewStripedTokenBucket(1e9, 1e9, time.Second, 64))
}
func BenchmarkAllowStriped256(b *testing.B) {
	benchAllow(b, NewStripedTokenBucket(1e9, 1e9, time.Second, 256))
}

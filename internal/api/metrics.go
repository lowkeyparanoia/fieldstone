package api

import (
	"net/http"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Real request metrics and a real activity log.
//
// These replace hardcoded dashboard numbers (TotalUsers: 45, RequestsPerMinute:
// 2340, StorageUsed: 150MB) and a fabricated activity feed that referenced
// users and collections which never existed. A dashboard that invents its own
// numbers is worse than one that shows nothing: it looks authoritative and
// misleads anybody reading it.
// ---------------------------------------------------------------------------

// requestCounter counts requests over a rolling window using second-granularity
// buckets, so "requests per minute" is a real measurement rather than a guess.
//
// A ring of 60 one-second buckets is used instead of a timestamp slice: memory
// is constant regardless of load, and expiry is O(1) rather than a scan.
type requestCounter struct {
	mu      sync.Mutex
	buckets [60]int64
	stamps  [60]int64 // unix second each bucket currently represents
}

func (c *requestCounter) observe(now time.Time) {
	sec := now.Unix()
	idx := sec % 60

	c.mu.Lock()
	defer c.mu.Unlock()
	// If this slot belongs to an older minute, it is stale: reset rather than
	// accumulate, otherwise counts from a minute ago leak into this one.
	if c.stamps[idx] != sec {
		c.stamps[idx] = sec
		c.buckets[idx] = 0
	}
	c.buckets[idx]++
}

// perMinute sums the buckets belonging to the last 60 seconds.
func (c *requestCounter) perMinute(now time.Time) int64 {
	cutoff := now.Unix() - 60
	var total int64

	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.buckets {
		if c.stamps[i] > cutoff {
			total += c.buckets[i]
		}
	}
	return total
}

// middleware counts every request that reaches the router.
func (c *requestCounter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.observe(time.Now())
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------------

// activityLog is a bounded in-memory ring of things that actually happened.
//
// Bounded deliberately: an unbounded slice is a memory leak with a long fuse.
// In-memory means it does not survive a restart, which is the honest trade for
// not adding a table and a migration. If durability is needed later this is the
// seam to swap for a real store.
type activityLog struct {
	mu    sync.RWMutex
	items []Activity
	max   int
}

func newActivityLog(max int) *activityLog {
	return &activityLog{items: make([]Activity, 0, max), max: max}
}

func (l *activityLog) record(a Activity) {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	// newest first, so the dashboard can take a prefix
	l.items = append([]Activity{a}, l.items...)
	if len(l.items) > l.max {
		l.items = l.items[:l.max]
	}
}

func (l *activityLog) recent(limit int) []Activity {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if limit <= 0 || limit > len(l.items) {
		limit = len(l.items)
	}
	out := make([]Activity, limit)
	copy(out, l.items[:limit])
	return out
}

// ---------------------------------------------------------------------------

// StorageReporter is an optional capability. Storage size is backend specific,
// so rather than widen the Backend interface for every implementation, the
// dashboard type-asserts for it and reports zero when unavailable. That keeps
// SQLite and memory backends honest instead of making them fake a number.
type StorageReporter interface {
	StorageBytes() (int64, error)
}

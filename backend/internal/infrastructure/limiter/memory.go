package limiter

import (
	"context"
	"sync"
	"time"
)

type MemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rate    int
	window  time.Duration
}

type tokenBucket struct {
	tokens    int
	lastReset time.Time
}

func NewMemoryLimiter(rate int, window time.Duration) *MemoryLimiter {
	ml := &MemoryLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    rate,
		window:  window,
	}

	go ml.cleanup()

	return ml
}

func (m *MemoryLimiter) Allow(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	bucket, exists := m.buckets[key]
	if !exists {
		m.buckets[key] = &tokenBucket{
			tokens:    m.rate - 1,
			lastReset: now,
		}
		return true, nil
	}

	if now.Sub(bucket.lastReset) >= m.window {
		bucket.tokens = m.rate - 1
		bucket.lastReset = now
		return true, nil
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true, nil
	}

	return false, nil
}

func (m *MemoryLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, bucket := range m.buckets {
			if now.Sub(bucket.lastReset) > m.window*2 {
				delete(m.buckets, key)
			}
		}
		m.mu.Unlock()
	}
}
package accrual

import (
	"context"
	"sync"
	"time"
)

type RateLimiter struct {
	mu    sync.Mutex
	until time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

func (rt *RateLimiter) Wait(ctx context.Context) {
	for {
		rt.mu.Lock()
		until := rt.until
		rt.mu.Unlock()

		wait := time.Until(until)

		if wait > 0 {
			timer := time.NewTimer(wait)

			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				timer.Stop()
			}
		} else {
			return
		}
	}
}

func (rt *RateLimiter) Block(interval time.Duration) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	newUntil := time.Now().Add(interval)

	if newUntil.After(rt.until) {
		rt.until = newUntil
	}
}

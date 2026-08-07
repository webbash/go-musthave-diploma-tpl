package accrual

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiterBlockAndWait(t *testing.T) {
	limiter := NewRateLimiter()
	limiter.Block(10 * time.Millisecond)

	start := time.Now()
	limiter.Wait(context.Background())
	if elapsed := time.Since(start); elapsed < 9*time.Millisecond {
		t.Fatalf("Wait() elapsed = %v, want at least 9ms", elapsed)
	}
}

func TestRateLimiterWaitContextCancel(t *testing.T) {
	limiter := NewRateLimiter()
	limiter.Block(time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	limiter.Wait(ctx)
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("Wait() elapsed = %v, want quick return on cancel", elapsed)
	}
}

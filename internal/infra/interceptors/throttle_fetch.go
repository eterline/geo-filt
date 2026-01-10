package interceptors

import (
	"context"
	"net/netip"
	"sync"
	"time"
)

type ThrottleWindowCallCtx struct {
	mu       sync.Mutex
	lastRun  time.Time
	interval time.Duration
}

func NewThrottleWindowCallCtx(interval time.Duration) *ThrottleWindowCallCtx {
	return &ThrottleWindowCallCtx{
		interval: interval,
	}
}

func (t *ThrottleWindowCallCtx) Do(ctx context.Context, key netip.Addr, fn func(context.Context, netip.Addr) (bool, error)) (bool, error) {

	for {
		t.mu.Lock()

		now := time.Now()
		next := t.lastRun.Add(t.interval)

		if now.Before(next) {
			wait := next.Sub(now)
			t.mu.Unlock()

			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return false, ctx.Err()
			}
			continue
		}

		t.lastRun = now
		t.mu.Unlock()
		break
	}

	return fn(ctx, key)
}

func WrapThrottleFetchAddr(interval time.Duration, job func(context.Context, netip.Addr) (bool, error)) func(context.Context, netip.Addr) (bool, error) {
	t := NewThrottleWindowCallCtx(interval)
	return func(ctx context.Context, key netip.Addr) (bool, error) {
		return t.Do(ctx, key, job)
	}
}

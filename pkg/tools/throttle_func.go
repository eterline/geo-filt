package tools

import (
	"sync"
	"time"
)

type ThrottleWindowCall struct {
	mu       sync.Mutex
	lastRun  time.Time
	interval time.Duration
}

func NewThrottleWindowCall(interval time.Duration) *ThrottleWindowCall {
	return &ThrottleWindowCall{
		interval: interval,
	}
}

func (r *ThrottleWindowCall) Do(fn func() (any, error)) (any, error) {
	for {
		r.mu.Lock()

		now := time.Now()
		next := r.lastRun.Add(r.interval)

		if now.Before(next) {
			wait := next.Sub(now)
			r.mu.Unlock()
			time.Sleep(wait)
			continue
		}

		r.lastRun = now
		r.mu.Unlock()
		break
	}

	return fn()
}

func WrapThrottleWindowCall(interval time.Duration, job func() (any, error)) func() (any, error) {
	t := NewThrottleWindowCall(interval)
	return func() (any, error) {
		return t.Do(job)
	}
}

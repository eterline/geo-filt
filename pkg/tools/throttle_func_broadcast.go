package tools

import (
	"sync"
	"time"
)

type ThrottleWindowCallBroadcast struct {
	mu   sync.Mutex
	cond *sync.Cond

	running bool
	lastRun time.Time

	result any
	err    error
}

func NewThrottleWindowCallBroadcast() *ThrottleWindowCallBroadcast {
	w := &ThrottleWindowCallBroadcast{}
	w.cond = sync.NewCond(&w.mu)
	return w
}

func (w *ThrottleWindowCallBroadcast) Do(window time.Duration, job func() (any, error)) (any, error) {
	w.mu.Lock()

	for w.running {
		w.cond.Wait()
	}

	now := time.Now()
	wait := w.lastRun.Add(window).Sub(now)

	if wait > 0 {

		timer := time.NewTimer(wait)
		w.running = true
		w.mu.Unlock()

		<-timer.C

		w.mu.Lock()
	} else {
		w.running = true
	}

	w.mu.Unlock()
	res, err := job()

	w.mu.Lock()
	w.lastRun = time.Now()
	w.result = res
	w.err = err
	w.running = false
	w.cond.Broadcast()
	w.mu.Unlock()

	return res, err
}

func WrapThrottleWindowCallBroadcast(window time.Duration, job func() (any, error)) func() (any, error) {
	br := NewThrottleWindowCallBroadcast()
	return func() (any, error) {
		return br.Do(window, job)
	}
}

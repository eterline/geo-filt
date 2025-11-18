package filter

import (
	"net/netip"
	"sync"
)

type RingIPCache struct {
	data  []netip.Addr
	head  int
	count int
	size  int
	mu    sync.RWMutex
}

func NewRingIPCache() *RingIPCache {
	return &RingIPCache{
		data: make([]netip.Addr, 128),
	}
}

func (rrb *RingIPCache) Exists(ip netip.Addr) bool {
	rrb.mu.RLock()
	defer rrb.mu.RUnlock()

	if rrb.count == 0 {
		return false
	}

	for i := 0; i < rrb.count; i++ {
		idx := (rrb.head + i) % rrb.size
		if rrb.data[idx] == ip {
			return true
		}
	}

	return false
}

func (rrb *RingIPCache) Remind(ip netip.Addr) {
	rrb.mu.Lock()
	defer rrb.mu.Unlock()

	if rrb.full() {
		rrb.data[rrb.head] = ip
		rrb.head = (rrb.head + 1) % rrb.size
		return
	}

	idx := (rrb.head + rrb.count) % rrb.size
	rrb.data[idx] = ip
	rrb.count++
}

func (rrb *RingIPCache) Flush() {
	rrb.mu.Lock()
	defer rrb.mu.Unlock()

	for i := 0; i < rrb.count; i++ {
		idx := (rrb.head + i) % rrb.size
		var dflt netip.Addr
		rrb.data[idx] = dflt
	}
	rrb.count = 0
	rrb.head = 0
}

func (rrb *RingIPCache) full() bool {
	return (rrb.count == rrb.size)
}

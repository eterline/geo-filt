package interceptors

import (
	"context"
	"net/netip"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type idempotentJob func(ctx context.Context, key netip.Addr) (bool, error)

// ===========================

type cacheEntry struct {
	allow bool
	exp   time.Time
}

func (e cacheEntry) Allowed() bool {
	return e.allow
}

// ===========================

type idempotentAllowTicketCache struct {
	cache map[netip.Addr]cacheEntry
	mu    sync.RWMutex
	sf    singleflight.Group
	job   idempotentJob
	ttl   time.Duration
}

func NewIPIdempotentAllowTicketCache(job idempotentJob, ttl time.Duration, cleanupInterval time.Duration) *idempotentAllowTicketCache {
	cache := &idempotentAllowTicketCache{
		cache: make(map[netip.Addr]cacheEntry),
		job:   job,
		ttl:   ttl,
	}

	go cache.cleanupLoop(cleanupInterval)

	return cache
}

func (c *idempotentAllowTicketCache) createCacheEntry(ticket bool) cacheEntry {
	return cacheEntry{
		allow: ticket,
		exp:   time.Now().Add(c.ttl),
	}
}

func (c *idempotentAllowTicketCache) GetAllowTicket(ctx context.Context, key netip.Addr) (bool, error) {

	c.mu.RLock()
	if ticket, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		return ticket.Allowed(), nil
	}
	c.mu.RUnlock()

	v, err, _ := c.sf.Do(key.String(),
		func() (any, error) {
			c.mu.RLock()
			if ticket, ok := c.cache[key]; ok {
				c.mu.RUnlock()
				return ticket.Allowed(), nil
			}
			c.mu.RUnlock()

			ticket, err := c.job(ctx, key)
			if err != nil {
				return nil, err
			}

			c.mu.Lock()
			c.cache[key] = c.createCacheEntry(ticket)
			c.mu.Unlock()

			return ticket, nil
		},
	)

	if err != nil {
		return false, err
	}

	return v.(bool), nil
}

func (c *idempotentAllowTicketCache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		c.mu.Lock()
		for ip, item := range c.cache {
			if now.After(item.exp) {
				delete(c.cache, ip)
			}
		}
		c.mu.Unlock()
	}
}

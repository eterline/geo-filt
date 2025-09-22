// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipmatch

import (
	"context"
	"errors"
	"net/netip"
	"sync"
)

type PoolMatcherIP struct {
	name string
	ctx  context.Context
	mu   sync.Mutex
	pool []netip.Prefix
}

func (m *PoolMatcherIP) Provider() string {
	return m.name
}

func (m *PoolMatcherIP) Match(ip netip.Addr) bool {

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.pool {
		if m.ctx.Err() != nil {
			return false
		}

		if p.Contains(ip) {
			return true
		}
	}

	return false
}

func (m *PoolMatcherIP) MatchParsed(s string) (bool, error) {
	ip, err := netip.ParseAddr(s)
	if err != nil {
		return false, errors.New("invalid ip address")
	}
	return m.Match(ip), nil
}

func (m *PoolMatcherIP) MustMatchParsed(s string) bool {
	ok, err := m.MatchParsed(s)
	if err != nil {
		panic(err)
	}
	return ok
}

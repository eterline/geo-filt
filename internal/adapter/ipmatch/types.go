// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipmatch

import (
	"errors"
	"net"
	"net/netip"
	"sync"
)

type PoolMatcherIP struct {
	name string
	mu   sync.Mutex
	pool []netip.Prefix
}

func (m *PoolMatcherIP) Provider() string {
	return m.name
}

func (m *PoolMatcherIP) Match(ip net.IP) bool {

	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	} else {
		ip = ip.To16()
	}

	if addr, ok := netip.AddrFromSlice(ip); ok {
		m.mu.Lock()
		defer m.mu.Unlock()

		for _, p := range m.pool {
			if p.Contains(addr) {
				return true
			}
		}
	}
	return false
}

func (m *PoolMatcherIP) MatchParsed(s string) (bool, error) {
	ip := net.ParseIP(s)
	if ip == nil {
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

// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package filter

import (
	"errors"
	"net/netip"
)

type MatchProvider interface {
	Provider() string
	Match(ip netip.Addr) bool
}

type IpFilterService struct {
	provideQueue []MatchProvider
}

func NewIpFilterService(q []MatchProvider) (*IpFilterService, error) {
	if q == nil && len(q) < 1 {
		return nil, errors.New("no IP sampling provider has been created")
	}

	return &IpFilterService{
		provideQueue: q,
	}, nil
}

func (ifs *IpFilterService) IsAllowed(ip netip.Addr) bool {
	for _, inst := range ifs.provideQueue {
		if inst.Match(ip) {
			return true
		}
	}
	return false
}

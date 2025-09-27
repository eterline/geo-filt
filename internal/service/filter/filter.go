// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package filter

import (
	"fmt"
	"net/netip"
	"os"
)

type MatchProvider interface {
	Provider() string
	Match(ip netip.Addr) bool
}

type IpFilterService struct {
	provideQueue []MatchProvider
}

func NewIpFilterService() *IpFilterService {
	return &IpFilterService{
		provideQueue: make([]MatchProvider, 0),
	}
}

// Add - adds MatchProvider object
func (ifs *IpFilterService) Add(mp MatchProvider) {
	if mp == nil {
		panic("match provider is nil")
	}
	os.Stdout.WriteString(fmt.Sprintf("match provider - added matcher: %s", mp.Provider()))
	ifs.provideQueue = append(ifs.provideQueue, mp)
}

func (ifs *IpFilterService) IsAllowed(ip netip.Addr) bool {
	for _, inst := range ifs.provideQueue {
		if inst.Match(ip) {
			return true
		}
	}
	return false
}

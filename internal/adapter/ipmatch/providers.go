// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipmatch

import (
	"context"
	"encoding/csv"
	"errors"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"go4.org/netipx"
)

type SubnetFileSelector map[int64]struct{}

func NewSubnetFileSelector(codesFile string, codes []string) (SubnetFileSelector, error) {
	file, err := resolvePath(codesFile, true)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	pool := map[int64]struct{}{}

	for i, code := range codes {
		codes[i] = strings.TrimSpace(strings.ToUpper(code))
	}

	for _, record := range records {
		if len(record) < 5 {
			continue
		}

		id, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			continue
		}

		for _, code := range codes {
			if record[4] == code {
				pool[id] = struct{}{}
				continue
			}
		}
	}

	return pool, nil
}

func (sfs SubnetFileSelector) SelectSubnets(subnetsFile []string) (*netipx.IPSet, error) {
	pool := &netipx.IPSetBuilder{}
	if len(subnetsFile) < 1 {
		return pool.IPSet()
	}

	for _, file := range subnetsFile {
		file, err := resolvePath(file, true)
		if err != nil {
			return nil, err
		}

		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		records, err := csv.NewReader(f).ReadAll()
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			if len(record) < 3 {
				continue
			}

			id, err := strconv.ParseInt(record[2], 10, 64)
			if err != nil {
				continue
			}

			if _, ok := sfs[id]; ok {
				pf, err := netip.ParsePrefix(record[0])
				if err != nil {
					continue
				}
				pool.AddPrefix(pf)
			}
		}
	}

	return pool.IPSet()
}

func NewMatcherGeoDB(ctx context.Context, countryFile string, subnetsFile []string, codes []string) (*PoolMatcherIP, error) {
	sl, err := NewSubnetFileSelector(countryFile, codes)
	if err != nil {
		return nil, err
	}

	pool, err := sl.SelectSubnets(subnetsFile)
	if err != nil {
		return nil, err
	}

	self := &PoolMatcherIP{
		name: "geodb",
		ctx:  ctx,
		pool: pool,
	}

	return self, nil
}

func NewMatcherDefinedSubnets(ctx context.Context, subnets []string) (*PoolMatcherIP, error) {
	if subnets == nil {
		return nil, errors.New("subnets is nil")
	}

	pool := &netipx.IPSetBuilder{}

	for _, s := range subnets {
		if p, err := netip.ParsePrefix(s); err == nil {
			pool.AddPrefix(p)
			continue
		}

		if ip, err := netip.ParseAddr(s); err == nil {
			var p netip.Prefix
			if ip.Is4() {
				p = netip.PrefixFrom(ip, 32)
			} else {
				p = netip.PrefixFrom(ip, 128)
			}
			pool.AddPrefix(p)
			continue
		}
	}

	set, err := pool.IPSet()
	if err != nil {
		return nil, err
	}

	self := &PoolMatcherIP{
		name: "defined",
		ctx:  ctx,
		pool: set,
	}

	return self, nil
}

/*
PrivateMatcher - matcher for private network IPs

	as RFC 1918 (IPv4 addresses), RFC 4193 (IPv6 addresses) and loopback IPs
*/
type PrivateMatcher struct{}

func NewPrivateMatcher() *PrivateMatcher {
	return &PrivateMatcher{}
}

func (m *PrivateMatcher) Match(ip netip.Addr) bool {
	return ip.IsLoopback() || ip.IsPrivate()

}

func (m *PrivateMatcher) Provider() string {
	return "private"
}

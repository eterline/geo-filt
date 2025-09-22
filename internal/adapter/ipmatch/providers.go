// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipmatch

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func selectCodeIDs(subnetsFile string, codes []string) (map[int64]struct{}, error) {
	file, err := resolvePath(subnetsFile, true)
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

func matchPrefixesById(subnetsFile string, idPool map[int64]struct{}) ([]netip.Prefix, error) {
	pool := []netip.Prefix{}
	if idPool == nil {
		return pool, nil
	}

	file, err := resolvePath(subnetsFile, true)
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

		if _, ok := idPool[id]; ok {
			pf, err := netip.ParsePrefix(record[0])
			if err != nil {
				continue
			}
			pool = append(pool, pf)
		}
	}

	return pool, nil
}

func NewMatcherGeoDB(ctx context.Context, countryFile, subnetsFile string, codes ...string) (*PoolMatcherIP, error) {
	idPool, err := selectCodeIDs(countryFile, codes)
	if err != nil {
		return nil, err
	}

	pool, err := matchPrefixesById(subnetsFile, idPool)
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

	pool := make([]netip.Prefix, 0, len(subnets))

	for _, s := range subnets {
		if p, err := netip.ParsePrefix(s); err == nil {
			pool = append(pool, p)
			continue
		}

		if ip, err := netip.ParseAddr(s); err == nil {
			var p netip.Prefix
			if ip.Is4() {
				p = netip.PrefixFrom(ip, 32)
			} else {
				p = netip.PrefixFrom(ip, 128)
			}
			pool = append(pool, p)
			continue
		}
	}

	self := &PoolMatcherIP{
		name: "defined",
		ctx:  ctx,
		pool: pool,
	}

	return self, nil
}

func resolvePath(input string, mustExist bool) (string, error) {
	if input == "" {
		return "", errors.New("empty path")
	}

	if strings.HasPrefix(input, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home dir: %w", err)
		}

		if input == "~" {
			input = home
		} else if strings.HasPrefix(input, "~/") || strings.HasPrefix(input, `~\`) {
			input = filepath.Join(home, input[2:])
		} else {
			return "", fmt.Errorf("unsupported ~user expansion: %q", input)
		}
	}

	abs, err := filepath.Abs(input)
	if err != nil {
		return "", fmt.Errorf("cannot make absolute path: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		abs = resolved
	}

	if mustExist {
		if _, err := os.Stat(abs); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("path does not exist: %s", abs)
			}
			return "", fmt.Errorf("stat error for %s: %w", abs, err)
		}
	}

	return abs, nil
}

type PrivateMatcher struct{}

func NewPrivateMatcher() *PrivateMatcher {
	return &PrivateMatcher{}
}

func (m *PrivateMatcher) Match(ip netip.Addr) bool {
	return ip.IsPrivate() || ip.IsLoopback()

}

func (m *PrivateMatcher) Provider() string {
	return "private"
}

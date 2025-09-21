// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipmatch

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/faireal/trojan-go/common/geodata"
	"github.com/v2fly/v2ray-core/v4/app/router"
)

// convertToNetip - converts v2ray-core router CIDRs to Go native netip.Prefix
func convertToNetip(cidrs []*router.CIDR) ([]netip.Prefix, error) {
	if cidrs == nil {
		return nil, errors.New("cidr list is nil")
	}

	pool := make([]netip.Prefix, 0, len(cidrs))

	for _, c := range cidrs {
		addr, ok := netip.AddrFromSlice(c.Ip) // c.Ip — []byte
		if !ok {
			continue
		}
		prefix := netip.PrefixFrom(addr, int(c.Prefix))
		pool = append(pool, prefix)
	}

	return pool, nil
}

func NewMatcherGeofileV2ray(filename, country string) (*PoolMatcherIP, error) {
	filename, err := ResolvePath(filename, false)
	if err != nil {
		return nil, err
	}

	loader := geodata.NewGeodataLoader()

	cidrs, err := loader.LoadIP(filename, country)
	if err != nil {
		return nil, fmt.Errorf("failed to load geofile: %w", err)
	}

	pool, err := convertToNetip(cidrs)
	if err != nil {
		return nil, fmt.Errorf("failed to load geofile: %w", err)
	}

	self := &PoolMatcherIP{
		name: "v2ray_geofile",
		pool: pool,
	}

	return self, nil
}

func NewMatcherDefinedSubnets(subnets []string) (*PoolMatcherIP, error) {
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
		pool: pool,
	}

	return self, nil
}

func ResolvePath(input string, mustExist bool) (string, error) {
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

// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipmatch

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"go4.org/netipx"
)

type PoolMatcherIP struct {
	name string
	ctx  context.Context
	pool *netipx.IPSet
}

func (m *PoolMatcherIP) Provider() string {
	return m.name
}

func (m *PoolMatcherIP) Match(ip netip.Addr) bool {
	if m.ctx.Err() != nil {
		return false
	}
	return m.pool.Contains(ip)
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

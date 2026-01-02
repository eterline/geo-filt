// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package filter

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"

	"github.com/eterline/geo-filt/internal/model"
)

type InterceptorFilter struct {
	interceptors []model.Interceptor
	log          *slog.Logger
}

func NewInterceptorFilter(log *slog.Logger, reg model.BuilderRegistry, cfgs []model.InterceptorConfig) (*InterceptorFilter, error) {
	interceptors := make([]model.Interceptor, len(cfgs))

	for i, cfg := range cfgs {
		interceptor, err := reg.BuildInterceptor(cfg)
		if err != nil {
			return nil, fmt.Errorf("cfg[%d]: failed to build interceptor: %w", i, err)
		}

		initLog := log
		if tag := interceptor.Tag(); tag != "" {
			initLog = initLog.With("tag", tag)
		}

		initLog.Info(
			"interceptor registered",
			"type", interceptor.Type(),
		)

		interceptors[i] = interceptor
	}

	return &InterceptorFilter{interceptors: interceptors}, nil
}

func (ifr *InterceptorFilter) IsAllowed(ctx context.Context, ip netip.Addr) (bool, error) {
	for _, inter := range ifr.interceptors {

		if err := ctx.Err(); err != nil {
			ifr.log.Debug("context canceled or timeout", "ip", ip, "err", err)
			return false, err
		}

		if inter.Match(ip) {
			ifr.log.Debug(
				"interceptor matched IP",
				"type", inter.Type(),
				"tag", inter.Tag(),
				"ip", ip.String(),
			)
			return true, nil
		}
	}

	return false, nil
}

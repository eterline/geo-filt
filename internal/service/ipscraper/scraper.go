// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipscraper

import (
	"net"
	"net/http"
	"net/netip"
)

type IpScraper struct {
	headerQueue []string
	req         *http.Request
}

func NewIpScraper(r *http.Request, h ...string) *IpScraper {
	return &IpScraper{
		headerQueue: h,
		req:         r,
	}
}

func (is *IpScraper) Scrape() (netip.Addr, bool) {
	ip, ok := is.headers()
	if ok {
		return ip, true
	}

	return is.remote()
}

func (is *IpScraper) headers() (netip.Addr, bool) {
	for _, wantedHeader := range is.headerQueue {
		if h := is.req.Header.Get(wantedHeader); h != "" {
			if ip, err := netip.ParseAddr(h); err != nil {
				return ip, true
			}
		}
	}
	return netip.Addr{}, false
}

func (is *IpScraper) remote() (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(is.req.RemoteAddr)
	if err != nil {
		return netip.Addr{}, false
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}

	return ip, true
}

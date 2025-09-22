// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipscraper

import (
	"net"
	"net/http"
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

func (is *IpScraper) Scrape() (net.IP, bool) {
	ip, ok := is.headers()
	if ok {
		return ip, true
	}

	return is.remote()
}

func (is *IpScraper) headers() (net.IP, bool) {
	for _, wantedHeader := range is.headerQueue {
		if h := is.req.Header.Get(wantedHeader); h != "" {
			if ip := net.ParseIP(h); ip != nil {
				return ip, true
			}
		}
	}
	return nil, false
}

func (is *IpScraper) remote() (net.IP, bool) {
	host, _, err := net.SplitHostPort(is.req.RemoteAddr)
	if err != nil {
		return nil, false
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, false
	}

	return ip, true
}

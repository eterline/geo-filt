// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipscraper

import (
	"net"
	"net/http"
)

type IpScraper struct {
	req *http.Request
}

func NewIpScraper(r *http.Request) *IpScraper {
	return &IpScraper{
		req: r,
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
	if h := is.req.Header.Get("X-Forwarded-For"); h != "" {
		if ip := net.ParseIP(h); ip != nil {
			return ip, true
		}
	}

	if h := is.req.Header.Get("True-Client-IP"); h != "" {
		if ip := net.ParseIP(h); ip != nil {
			return ip, true
		}
	}

	if h := is.req.Header.Get("X-Real-IP"); h != "" {
		if ip := net.ParseIP(h); ip != nil {
			return ip, true
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

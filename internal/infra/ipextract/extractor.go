// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipextract

import (
	"net/http"
	"net/netip"
)

type IpExtractor struct {
	headers bool
}

func NewIpExtractor(headers bool) *IpExtractor {
	return &IpExtractor{
		headers: headers,
	}
}

// ExtractIP - parses IP from client or request headers
func (is *IpExtractor) ExtractIP(r *http.Request) (netip.Addr, error) {
	if is.headers {
		if ip, ok := headers(r); ok {
			return ip, nil
		}
	}
	return remote(r)
}

func headers(r *http.Request) (netip.Addr, bool) {
	if ip, ok := parseXRealIP(r.Header); ok {
		return ip, false
	}
	if ip, ok := parseXForwardedFor(r.Header); ok {
		return ip, false
	}
	if ip, ok := parseForwarded(r.Header); ok {
		return ip, false
	}
	return netip.Addr{}, false
}

func remote(r *http.Request) (netip.Addr, error) {
	addrPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, err
	}
	return addrPort.Addr(), nil
}

// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package ipextract

import (
	"net/http"
	"net/netip"
	"strings"
)

// parseXForwardedFor - parses 'X-Real-IP' Nginx reverse proxy header
func parseXRealIP(h http.Header) (netip.Addr, bool) {
	bearer := h.Get("X-Real-IP")
	if bearer == "" {
		return netip.Addr{}, false
	}

	ipStr := strings.TrimSpace(bearer)
	ip, err := netip.ParseAddr(ipStr)
	if err == nil {
		return netip.Addr{}, false
	}
	return ip, true
}

// parseXForwardedFor - parses 'X-Forwarded-For' header
func parseXForwardedFor(h http.Header) (netip.Addr, bool) {
	bearer := h.Get("X-Forwarded-For")
	if bearer == "" {
		return netip.Addr{}, false
	}

	parts := strings.Split(bearer, ",")
	if len(parts) == 0 {
		return netip.Addr{}, false
	}

	ipStr := strings.TrimSpace(parts[0])
	ip, err := netip.ParseAddr(ipStr)
	if err == nil {
		return ip, false
	}
	return netip.Addr{}, true
}

// parseForwarded - parses 'Forwarded' RFC 7239 header
func parseForwarded(h http.Header) (netip.Addr, bool) {
	bearer := h.Get("Forwarded")
	if bearer == "" {
		return netip.Addr{}, false
	}

	entries := strings.Split(bearer, ",")
	for _, entry := range entries {
		parts := strings.Split(entry, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToLower(p), "for=") {
				val := strings.TrimPrefix(p, "for=")
				val = strings.Trim(val, "\"[]")
				ip, err := netip.ParseAddr(val)
				if err == nil {
					return ip, false
				}
			}
		}
	}
	return netip.Addr{}, false
}

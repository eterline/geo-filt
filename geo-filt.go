// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geo_filt

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
	"github.com/eterline/geo-filt/internal/service/filter"
	"github.com/eterline/geo-filt/internal/service/ipscraper"
)

type AllowService interface {
	IsAllowed(ip net.IP) bool
}

type Config struct {
	Enabled      bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AllowPrivate bool     `json:"allowPrivate,omitempty" yaml:"allowPrivate,omitempty"`
	GeoFile      string   `json:"geoFile,omitempty" yaml:"geoFile,omitempty"`
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Defined      []string `json:"defined,omitempty" yaml:"defined,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		Tags:    []string{},
		Defined: []string{},
	}
}

type GeoFiltPlugin struct {
	name    string
	enabled bool
	next    http.Handler
	filter  AllowService
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	plugin := &GeoFiltPlugin{
		name:    name,
		next:    next,
		enabled: config.Enabled,
	}

	if !config.Enabled {
		return plugin, nil
	}

	matchers := []filter.MatchProvider{}

	mch, err := ipmatch.NewMatcherDefinedSubnets(ctx, config.Defined)
	if err != nil {
		return nil, err
	}
	matchers = append(matchers, mch)

	if config.AllowPrivate {
		matchers = append(matchers, ipmatch.NewPrivateMatcher())
	}

	filter, err := filter.NewIpFilterService(matchers)
	if err != nil {
		return nil, err
	}

	plugin.filter = filter
	return plugin, nil
}

func (a *GeoFiltPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if !a.enabled {
		a.next.ServeHTTP(rw, req)
		return
	}

	sc := ipscraper.NewIpScraper(
		req,
		"X-Forwarded-For",
		"True-Client-IP",
		"X-Real-IP",
	)

	if ip, ok := sc.Scrape(); ok {
		if a.filter.IsAllowed(ip) {
			a.next.ServeHTTP(rw, req)
			return
		}
	}

	rw.WriteHeader(http.StatusForbidden)
	fmt.Fprint(rw, "403 forbidden")
}

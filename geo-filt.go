// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geofilt

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/eterline/geo-filt/internal/service/filter"
	"github.com/eterline/geo-filt/internal/service/ipscraper"
)

type AllowService interface {
	IsAllowed(ip net.IP) bool
}

type Config struct {
	GeoFile string   `json:"geofile,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Defined []string `json:"defined,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		Tags:    []string{},
		Defined: []string{},
	}
}

type GeoFiltPlugin struct {
	name   string
	next   http.Handler
	filter AllowService
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	s, err := filter.NewIpFilterServiceMock()
	if err != nil {
		return nil, err
	}

	return &GeoFiltPlugin{
		name:   name,
		next:   next,
		filter: s,
	}, nil
}

func (a *GeoFiltPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	sc := ipscraper.NewIpScraper(req)

	if ip, ok := sc.Scrape(); ok {
		if a.filter.IsAllowed(ip) {
			a.next.ServeHTTP(rw, req)
			return
		}
	}

	rw.WriteHeader(http.StatusForbidden)
	fmt.Fprint(rw, "403 forbidden")
}

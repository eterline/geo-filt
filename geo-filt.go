// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geo_filt

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"os"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
	"github.com/eterline/geo-filt/internal/service/filter"
	"github.com/eterline/geo-filt/internal/service/ipscraper"
)

type AllowService interface {
	IsAllowed(ip netip.Addr) bool
}

type ExtractorIP interface {
	ExtractIP(r *http.Request) (netip.Addr, bool)
}

// ===========================

// Config - plugin basic configuration
type Config struct {
	Enabled      bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AllowPrivate bool     `json:"allowPrivate,omitempty" yaml:"allowPrivate,omitempty"`
	HeaderBearer bool     `json:"headerBearer,omitempty" yaml:"headerBearer,omitempty"`
	CodeFile     string   `json:"codeFile,omitempty" yaml:"codeFile,omitempty"`
	GeoFile      string   `json:"geoFile,omitempty" yaml:"geoFile,omitempty"`
	GeoFile6     string   `json:"geoFile6,omitempty" yaml:"geoFile6,omitempty"`
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Defined      []string `json:"defined,omitempty" yaml:"defined,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		Enabled:      false,
		AllowPrivate: false,
		HeaderBearer: false,
		CodeFile:     "",
		GeoFile:      "",
		GeoFile6:     "",
		Tags:         []string{},
		Defined:      []string{},
	}
}

/*
	geoConfExists - tests available geo config strings.

okV4 - IPv4 file exists. okV6 - IPv6 file exists.
*/
func (c Config) geoConfExists() (okV4, okV6 bool) {
	if len(c.Tags) > 0 {
		return (c.GeoFile != ""), (c.GeoFile6 != "")
	}
	return false, false
}

// ===========================

type GeoFiltPlugin struct {
	name      string
	enabled   bool
	next      http.Handler
	filter    AllowService
	ipExtract ExtractorIP
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	extractor := ipscraper.NewIpExtractor(config.HeaderBearer)
	plugin := &GeoFiltPlugin{
		name:      name,
		next:      next,
		enabled:   config.Enabled,
		ipExtract: extractor,
	}

	// if disabled, plugin will pass request in any case
	if !config.Enabled {
		fmt.Fprint(os.Stdout, "geo-filt - disabled. Skip configuration\n")
		return plugin, nil
	}

	// init match queue
	queue := []filter.MatchProvider{}

	fmt.Fprint(os.Stdout, "geo-filt - starting init configuration\n")

	// default allow for private network IPs
	// as RFC 1918 (IPv4 addresses) and RFC 4193 (IPv6 addresses)
	// includes loopback IPs
	if config.AllowPrivate {
		queue = append(queue, ipmatch.NewPrivateMatcher())
	}

	// allow defined in config subnets and IPs (look at Config.Defined)
	mch, err := ipmatch.NewMatcherDefinedSubnets(ctx, config.Defined)
	if err != nil {
		return nil, err
	}
	queue = append(queue, mch)

	v4, v6 := config.geoConfExists()

	if v4 {
		mch, err := ipmatch.NewMatcherGeoDB(ctx, config.CodeFile, config.GeoFile, config.Tags)
		if err != nil {
			return nil, err
		}
		queue = append(queue, mch)
	}

	if v6 {
		mch, err := ipmatch.NewMatcherGeoDB(ctx, config.CodeFile, config.GeoFile6, config.Tags)
		if err != nil {
			return nil, err
		}
		queue = append(queue, mch)
	}

	filter, err := filter.NewIpFilterService(queue)
	if err != nil {
		return nil, err
	}

	for _, matcher := range queue {
		fmt.Fprintf(os.Stdout, "geo-filt - include '%s' matcher\n", matcher.Provider())
	}

	// set filter service to plugin
	plugin.filter = filter
	return plugin, nil
}

func (plugin *GeoFiltPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if !plugin.enabled {
		// transparent use
		plugin.next.ServeHTTP(rw, req)
		return
	}

	if ip, ok := plugin.ipExtract.ExtractIP(req); ok {
		if plugin.filter.IsAllowed(ip) {
			plugin.next.ServeHTTP(rw, req)
			return
		}
	}

	http.Error(rw, "403 Forbidden - Invalid request region", http.StatusForbidden)
}

// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geo_filt

import (
	"context"
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
	GeoFile      []string `json:"geoFile,omitempty" yaml:"geoFile,omitempty"`
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Defined      []string `json:"defined,omitempty" yaml:"defined,omitempty"`
}

func CreateConfig() *Config {
	return &Config{
		Enabled:      false,
		AllowPrivate: false,
		HeaderBearer: false,
		CodeFile:     "",
		GeoFile:      []string{},
		Tags:         []string{},
		Defined:      []string{},
	}
}

// geoConfExists - tests available geo config strings.
func (c Config) geoConfExists() bool {
	return (len(c.Tags) > 0) &&
		(len(c.GeoFile) > 0) &&
		(c.CodeFile != "")
}

// definedExists - tests available defined strings.
func (c Config) definedExists() bool {
	return len(c.Defined) > 0
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
	os.Stdout.WriteString("geo-filt - starting init configuration")
	filter := filter.NewIpFilterService()
	plugin := &GeoFiltPlugin{
		name:    name,
		next:    next,
		enabled: config.Enabled,
		// set filter service to plugin
		filter: filter,
		// set extracting IP service to plugin
		ipExtract: ipscraper.NewIpExtractor(config.HeaderBearer),
	}

	// if disabled, plugin will pass request in any case
	if !config.Enabled {
		os.Stdout.WriteString("geo-filt - disabled. Skip configuration")
		return plugin, nil
	}

	// allow defined in config subnets and IPs (look at Config.Defined)
	if config.definedExists() {
		mch, err := ipmatch.NewMatcherDefinedSubnets(ctx, config.Defined)
		if err != nil {
			return nil, err
		}
		filter.Add(mch)
	}

	// default allow for private network IPs
	// as RFC 1918 (IPv4 addresses) and RFC 4193 (IPv6 addresses)
	// includes loopback IPs
	if config.AllowPrivate {
		mch := ipmatch.NewPrivateMatcher()
		filter.Add(mch)
	}

	// allow subnets from GeoDB
	if config.geoConfExists() {
		mch, err := ipmatch.NewMatcherGeoDB(ctx, config.CodeFile, config.GeoFile, config.Tags)
		if err != nil {
			return nil, err
		}
		filter.Add(mch)
	}

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

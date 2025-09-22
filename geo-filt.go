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

/*
	geoConfExists - tests available geo config strings.

okV4 - IPv4 file exists. okV6 - IPv6 file exists.
*/
func (c Config) geoConfExists() bool {
	return (len(c.Tags) > 0) && (len(c.GeoFile) > 0)
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

	plugin := &GeoFiltPlugin{
		name:    name,
		next:    next,
		enabled: config.Enabled,
	}

	// if disabled, plugin will pass request in any case
	if !config.Enabled {
		os.Stdout.WriteString("geo-filt - disabled. Skip configuration")
		return plugin, nil
	}

	// init match queue
	queue := []filter.MatchProvider{}
	os.Stdout.WriteString("geo-filt - starting init configuration")

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

	if config.geoConfExists() {
		mch, err = ipmatch.NewMatcherGeoDB(ctx, config.CodeFile, config.GeoFile, config.Tags)
		if err != nil {
			return nil, err
		}
		queue = append(queue, mch)
	}

	fl, err := filter.NewIpFilterService(queue)
	if err != nil {
		return nil, err
	}

	// set filter service to plugin
	plugin.filter = fl
	// set extracting IP service to plugin
	plugin.ipExtract = ipscraper.NewIpExtractor(config.HeaderBearer)

	for _, matcher := range queue {
		msg := fmt.Sprintf("geo-filt - include '%s' matcher", matcher.Provider())
		os.Stdout.WriteString(msg)
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

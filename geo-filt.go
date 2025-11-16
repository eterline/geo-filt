// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geo_filt

import (
	"context"
	"log/slog"
	"net/http"
	"net/netip"

	"github.com/eterline/geo-filt/internal/adapter/ipmatch"
	"github.com/eterline/geo-filt/internal/service/filter"
	"github.com/eterline/geo-filt/internal/service/ipscraper"
)

type FilterCache interface {
	Exists(ip netip.Addr) bool
	Remind(ip netip.Addr)
	Flush()
}

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
	Cache        bool     `json:"cache,omitempty" yaml:"cache,omitempty"`
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
		Cache:        false,
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
	cache     FilterCache
	ipExtract ExtractorIP
	log       *slog.Logger
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	log := slog.With("addon", "geo-filt")
	log.Info("staring")

	filterSrvc := filter.NewIpFilterService()
	plugin := &GeoFiltPlugin{
		name:    name,
		next:    next,
		enabled: config.Enabled,
		// set filter service to plugin
		filter: filterSrvc,
		// set extracting IP service to plugin
		ipExtract: ipscraper.NewIpExtractor(config.HeaderBearer),
		log:       log,
	}

	// if disabled, plugin will pass request in any case
	if !config.Enabled {
		log.Info("skip initialization", "reason", "addon disabled")
		return plugin, nil
	}

	if config.Cache {
		log.Info("ip cache enabled")
		plugin.cache = filter.NewRingBufferIPCache()
	}

	// allow defined in config subnets and IPs (look at Config.Defined)
	if config.definedExists() {
		log := log.With("subnets", config.Defined)

		mch, err := ipmatch.NewMatcherDefinedSubnets(ctx, config.Defined)
		if err != nil {
			log.Error("failed add predefined subnets", "error", err.Error())
			return nil, err
		}

		log.Info("predefined allowed subnets setup")
		filterSrvc.Add(mch)
	}

	// default allow for private network IPs
	// as RFC 1918 (IPv4 addresses) and RFC 4193 (IPv6 addresses)
	// includes loopback IPs
	if config.AllowPrivate {
		log.Info("private subnets allowed")
		mch := ipmatch.NewPrivateMatcher()
		filterSrvc.Add(mch)
	}

	// allow subnets from GeoDB
	if config.geoConfExists() {
		log := log.
			With("code_file", config.CodeFile).
			With("geodb_files", config.GeoFile).
			With("geo_tags", config.Tags)

		mch, err := ipmatch.NewMatcherGeoDB(ctx, config.CodeFile, config.GeoFile, config.Tags)
		if err != nil {
			log.Error("failed to config geodata", "error", err.Error())
			return nil, err
		}

		// log.Info("geodata filter initialized")
		filterSrvc.Add(mch)
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
		if plugin.cache != nil && plugin.cache.Exists(ip) {
			plugin.next.ServeHTTP(rw, req)
			return
		}

		if plugin.filter.IsAllowed(ip) {
			if plugin.cache != nil {
				plugin.cache.Remind(ip)
			}

			plugin.next.ServeHTTP(rw, req)
			return
		}

		plugin.log.Info("request failed filtering", "addr", ip.String())
	}

	http.Error(rw, "FORBIDDEN - invalid request region", http.StatusForbidden)
}

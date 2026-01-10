// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geo_filt

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/eterline/geo-filt/internal/infra/forbidden"
	"github.com/eterline/geo-filt/internal/infra/interceptors"
	"github.com/eterline/geo-filt/internal/infra/ipextract"
	"github.com/eterline/geo-filt/internal/infra/log"
	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/internal/service/filter"
)

// Config the plugin configuration.
// Config defines the configuration for the GeoFilt plugin.
type Config struct {
	Enabled      bool                      `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	HeaderIP     bool                      `json:"header_ip,omitempty" yaml:"header_ip,omitempty"`
	LogLevel     string                    `json:"log,omitempty" yaml:"log,omitempty"`
	Response     model.ForbiddenConfig     `json:"response,omitempty" yaml:"response,omitempty"`
	Interceptors []model.InterceptorConfig `json:"interceptors" yaml:"interceptors"`
}

// CreateConfig returns a new default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Enabled:      false,
		HeaderIP:     false,
		LogLevel:     "info",
		Response:     model.ForbiddenConfig{},
		Interceptors: make([]model.InterceptorConfig, 0),
	}
}

// GeoFiltPlugin represents the GeoIP filtering plugin.
type GeoFiltPlugin struct {
	name          string
	enabled       bool
	next          http.Handler
	extract       model.ExtractorIP
	filter        model.IPFilter
	forbiddWriter model.ResponseForbiddenWriter
	log           *slog.Logger
}

// New creates a new GeoFiltPlugin instance with the given configuration.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	log := log.NewLogger(config.LogLevel, false).With(
		"plugin", "geo-filt",
		"plugin_name", name,
	)

	// If the plugin is disabled, just return a no-op plugin.
	if !config.Enabled {
		log.Warn("plugin disabled")
		plugin := &GeoFiltPlugin{
			name:    name,
			enabled: false,
			next:    next,
			log:     log,
		}
		return plugin, nil
	}

	log.Info("starting plugin")

	// Ensure `next` handler is not nil
	if next == nil {
		log.Warn("nil next handler, using NotFoundHandler()")
		next = http.NotFoundHandler()
	}

	// Create IP extractor
	ipExtract := ipextract.NewIpExtractor(config.HeaderIP)
	log.Info("setup IP extractor", "lookup_header_ip", config.HeaderIP)

	// Create forbidden response writer
	forbiddWriter, err := forbidden.InitForbiddenWriter(config.Response.Type, config.Response.Content)
	if err != nil {
		log.Error("error setup forbidden response adapter", "error", err)
		return nil, err
	}
	log.Info("setup forbidden response adapter", "type", config.Response.Type)

	// Initialize IP filter with configured interceptors
	ipFilter, err := filter.NewInterceptorFilter(log, interceptors.SetupRegistry, config.Interceptors)
	if err != nil {
		log.Error("failed to start interceptor filter", "error", err)
		return nil, err
	}

	// Create plugin instance
	plugin := &GeoFiltPlugin{
		name:          name,
		enabled:       config.Enabled,
		next:          next,
		extract:       ipExtract,
		filter:        ipFilter,
		forbiddWriter: forbiddWriter,
		log:           log,
	}

	return plugin, nil
}

// ServeHTTP handles the incoming HTTP request, performing IP filtering.
func (p *GeoFiltPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If the plugin is disabled, just pass the request to the next handler.
	if !p.enabled {
		p.log.Debug("plugin disabled, skip filtering", "ip", r.RemoteAddr)
		p.next.ServeHTTP(w, r)
		return
	}

	// Extract client IP
	clientIP, err := p.extract.ExtractIP(r)
	if err != nil {
		p.log.Error("error extracting request IP", "error", err)
		p.forbiddWriter.ResponseForbidden(w)
		return
	}

	// Check if the IP is allowed
	allowed, err := p.filter.IsAllowed(r.Context(), clientIP)
	if err != nil {
		// Ignore context cancellations
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			p.log.Debug("request context canceled", "ip", clientIP, "error", err)
			return
		}

		// Log other errors
		p.log.Error("error checking IP filter", "ip", clientIP, "error", err)
		p.forbiddWriter.ResponseForbidden(w)
		return
	}

	// Block request if IP is not allowed
	if !allowed {
		p.log.Debug("request blocked by IP filter", "ip", clientIP)
		p.forbiddWriter.ResponseForbidden(w)
		return
	}

	// Request allowed
	p.log.Debug("request allowed", "ip", clientIP)
	p.next.ServeHTTP(w, r)
}

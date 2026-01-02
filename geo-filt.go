// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geo_filt

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/eterline/geo-filt/internal/infra/interceptors"
	"github.com/eterline/geo-filt/internal/infra/ipextract"
	"github.com/eterline/geo-filt/internal/interface/forbidden"
	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/internal/service/filter"
)

// Config - plugin basic configuration
type Config struct {
	Enabled   bool                      `json:"enabled" yaml:"enabled"`
	HeaderIP  bool                      `json:"header_ip" yaml:"header_ip"`
	Response  model.ForbiddenConfig     `json:"response" yaml:"response"`
	Intercept []model.InterceptorConfig `json:"intercept" yaml:"intercept"`
}

func CreateConfig() *Config {
	return &Config{
		Enabled:   false,
		Intercept: make([]model.InterceptorConfig, 0),
	}
}

// ===========================

type GeoFiltPlugin struct {
	name          string
	enabled       bool
	next          http.Handler
	extract       model.ExtractorIP
	filter        model.IPFilter
	forbiddWriter model.ResponseForbiddenWriter
	log           *slog.Logger
}

// ===========================

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	log := slog.With(
		"plugin", "geo-filt",
		"plugin_name", name,
	)

	log.Info("staring plugin")

	log.Info("setup ip extracting adapter", "lookup_header_ip", config.HeaderIP)
	extractor := ipextract.NewIpExtractor(config.HeaderIP)

	ipFilter, err := filter.NewInterceptorFilter(log, interceptors.SetupRegistry, config.Intercept)
	if err != nil {
		log.Error(
			"failed to start interceptor filter",
			"error", err,
		)
		return nil, err
	}

	var forbiddWriter model.ResponseForbiddenWriter
	switch config.Response.Type {
	case "html":
		forbiddWriter = forbidden.NewHTMLWriter(config.Response.Content)
	default:
		forbiddWriter = forbidden.NewPlainWriter(config.Response.Content)
	}

	plugin := &GeoFiltPlugin{
		name:          name,
		enabled:       config.Enabled,
		next:          next,
		extract:       extractor,
		filter:        ipFilter,
		forbiddWriter: forbiddWriter,
		log:           log,
	}

	return plugin, nil
}

func (p *GeoFiltPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !p.enabled {
		p.next.ServeHTTP(w, r)
		return
	}

	clientIP := p.extract.ExtractIP(r)

	allowed, err := p.filter.IsAllowed(r.Context(), clientIP)
	if err != nil {
		p.log.Error("error checking IP filter", "ip", clientIP, "err", err)
		p.forbiddWriter.ResponseForbidden(w)
		return
	}

	if !allowed {
		p.log.Debug("request blocked by IP filter", "ip", clientIP)
		p.forbiddWriter.ResponseForbidden(w)
		return
	}

	p.log.Debug("request allowed", "ip", clientIP)
	p.next.ServeHTTP(w, r)
}

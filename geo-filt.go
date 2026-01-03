// Copyright (c) 2025 EterLine (Andrew)
// This file is part of geo-filt.
// Licensed under the GNU AFFERO GENERAL PUBLIC LICENSE. See the LICENSE file for details.

package geo_filt

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/eterline/geo-filt/internal/infra/forbidden"
	"github.com/eterline/geo-filt/internal/infra/interceptors"
	"github.com/eterline/geo-filt/internal/infra/ipextract"
	"github.com/eterline/geo-filt/internal/model"
	"github.com/eterline/geo-filt/internal/service/filter"
)

// Config - plugin basic configuration
type Config struct {
	Enabled      bool                      `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	HeaderIP     bool                      `json:"header_ip,omitempty" yaml:"header_ip,omitempty"`
	Response     model.ForbiddenConfig     `json:"response,omitempty" yaml:"response,omitempty"`
	Interceptors []model.InterceptorConfig `json:"interceptors" yaml:"interceptors"`
}

func CreateConfig() *Config {
	return &Config{
		Enabled:      false,
		Interceptors: make([]model.InterceptorConfig, 0),
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

	// ======

	ipFilter, err := filter.NewInterceptorFilter(
		log,
		interceptors.SetupRegistry,
		config.Interceptors,
	)

	if err != nil {
		log.Error(
			"failed to start interceptor filter",
			"error", err,
		)
		return nil, err
	}

	ipExtract := ipextract.NewIpExtractor(config.HeaderIP)
	log.Info(
		"setup ip extracting adapter",
		"lookup_header_ip", config.HeaderIP,
	)

	r := config.Response
	forbiddWriter := forbidden.InitForbiddenWriter(r.Type, r.Content)
	log.Info(
		"setup response adapter",
		"content", r.Content,
	)

	// ======

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

	if p.next != nil {
		p.log.Debug("request allowed", "ip", clientIP)
		p.next.ServeHTTP(w, r)
		return
	}

	p.log.Error("next handler is nil")
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
